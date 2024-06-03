package kubelet

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"minik8s/internal/apiobject" // Ensure correct import path
	"strings"

	"minik8s/utils"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
)

type RuntimeManager struct {
	DockerClient *DockerClient
}

// NewRuntimeManager initializes and returns a new RuntimeManager
func NewRuntimeManager(dockerClient *DockerClient) *RuntimeManager {
	return &RuntimeManager{DockerClient: dockerClient}
}

func calculateCPUPercentUnix(previous types.Stats) float64 {
	cpuDelta := float64(previous.CPUStats.CPUUsage.TotalUsage) - float64(previous.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(previous.CPUStats.SystemUsage) - float64(previous.PreCPUStats.SystemUsage)
	onlineCPUs := float64(previous.CPUStats.OnlineCPUs)
	if onlineCPUs == 0.0 {
		onlineCPUs = float64(len(previous.CPUStats.CPUUsage.PercpuUsage))
	}
	cpuPercent := (cpuDelta / systemDelta) * onlineCPUs
	return cpuPercent
}

func calculateMemPercentUnix(previous types.Stats) float64 {
	memPercent := float64(previous.MemoryStats.Usage) / float64(previous.MemoryStats.Limit)
	return memPercent
}

func (r *RuntimeManager) GetContainerResource(containerID string) (float64, float64, error) {
	ctx := context.Background()
	stats, err := r.DockerClient.Client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get container stats for %s: %v", containerID, err)
	}
	defer stats.Body.Close()

	var stat types.Stats
	if err := json.NewDecoder(stats.Body).Decode(&stat); err != nil {
		return 0, 0, fmt.Errorf("failed to decode container stats for %s: %v", containerID, err)
	}

	cpuPercent := calculateCPUPercentUnix(stat)
	memoryPercent := calculateMemPercentUnix(stat)
	return cpuPercent, memoryPercent, nil
}

// CreatePod creates a pod with the specified configuration
func (r *RuntimeManager) CreatePod(pod *apiobject.PodStore) error {
	ctx := context.Background()
	images, _ := r.DockerClient.Client.ImageList(ctx, image.ListOptions{})

	var pauseID string
	var err error

	if pauseID, err = r.createPauseContainer(images, ctx, pod); err != nil {
		return fmt.Errorf("failed to create pause container: %v", err)
	}

	for _, containerSpec := range pod.Spec.Containers {
		containerID, _ := r.createContainerWithLabel(images, ctx, pauseID, pod.Metadata, containerSpec)
		if containerID != "" {
			cpuPercent, memoryPercent, _ := r.GetContainerResource(containerID)
			pod.Status.CpuPercent += cpuPercent
			pod.Status.MemPercent += memoryPercent
		}
	}

	return nil
}

// creae pause container
func (r *RuntimeManager) createPauseContainer(images []image.Summary, ctx context.Context, pod *apiobject.PodStore) (string, error) {
	pauseImage := "registry.aliyuncs.com/google_containers/pause:3.6"
	if !r.DockerClient.ImageExists(images, pauseImage) {
		if err := r.DockerClient.PullImage(pauseImage); err != nil {
			log.Printf("Failed to pull pause image %s: %v", pauseImage, err)
			return "", err
		}
	}
	// Create or get the pause container ID
	pauseContainerName := pod.Metadata.Name + "_pause"
	labels := map[string]string{"pod_uid": pod.Metadata.UUID}

	var pauseID string
	var err error

	if !r.DockerClient.ContainerExists(pauseContainerName) {
		// Create the pause container if not present
		pauseCntConfig := container.Config{
			Image:   pauseImage,
			Volumes: nil,
			Env:     nil,
			Labels:  labels,
		}
		pauseHostConfig := container.HostConfig{IpcMode: "shareable"}
		pauseResp, err := r.DockerClient.Client.ContainerCreate(ctx, &pauseCntConfig, &pauseHostConfig, nil, nil, pauseContainerName)
		if err != nil {
			log.Printf("Failed to create pause container: %v", err)
			return "", err
		}
		pauseID = pauseResp.ID

		// Start the pause container
		if err := r.DockerClient.Client.ContainerStart(ctx, pauseID, container.StartOptions{}); err != nil {
			log.Printf("Failed to start pause container: %v", err)
			return "", err
		}

	} else {
		// Get the existing pause container ID
		pauseID, err = r.DockerClient.GetContainerIDByName(pauseContainerName)
		if err != nil {
			log.Printf("Failed to get pause container ID: %v", err)
			return "", err
		}
	}

	// [Weave网络] 为pause容器添加网络
	if pod.Status.PodIP == "" {
		res, err := utils.AttachContainer(pauseID)
		if err != nil {
			log.Fatal("Pause Container", err.Error()+res)
			return "", err
		}
		pod.Status.PodIP = strings.TrimSuffix(res, "\n")
		log.Printf("WeaveAttach res %v", res)
	}
	return pauseID, nil
}

func (r *RuntimeManager) createContainerWithLabel(images []image.Summary, ctx context.Context, pauseID string, podMeta apiobject.Metadata, containerSpec apiobject.Container) (string, error) {
	// Check and pull the container image if not present
	if !r.DockerClient.ImageExists(images, containerSpec.Image) {
		if err := r.DockerClient.PullImage(containerSpec.Image); err != nil {
			return "", fmt.Errorf("failed to pull image %s: %v", containerSpec.Image, err)
		}
	}

	containerName := fmt.Sprintf("%s_%s", podMeta.Name, containerSpec.Name)
	if !r.DockerClient.ContainerExists(containerName) {
		labels := map[string]string{"pod_uid": podMeta.UUID}
		pauseRef := "container:" + pauseID

		contConf := container.Config{
			Image:      containerSpec.Image,
			Entrypoint: containerSpec.Command,
			Cmd:        containerSpec.Args,
			Labels:     labels,
		}
		contRes := container.Resources{}
		if containerSpec.Resources.Limits.Memory != "" && containerSpec.Resources.Limits.Cpu != "" {
			mem, _ := MemoryToBytes(containerSpec.Resources.Limits.Memory)
			cpuPer, _ := CpuToMillicores(containerSpec.Resources.Limits.Cpu)
			contRes = container.Resources{
				Memory:   mem,
				NanoCPUs: cpuPer,
			}
		}
		hostConf := container.HostConfig{
			PidMode:     container.PidMode(pauseRef),
			IpcMode:     container.IpcMode(pauseRef),
			NetworkMode: container.NetworkMode(pauseRef),
			Resources:   contRes,
		}
		// Create the application container if not present
		resp, err := r.DockerClient.Client.ContainerCreate(ctx, &contConf, &hostConf, &network.NetworkingConfig{}, nil, containerName)
		if err != nil {
			return "", fmt.Errorf("failed to create container %s: %v", containerSpec.Image, err)
		}

		// Start the application container
		if err := r.DockerClient.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
			return "", fmt.Errorf("failed to start container %s: %v", resp.ID, err)
		}
		fmt.Printf("Started container %s with ID %s\n", containerSpec.Image, resp.ID)
		return resp.ID, nil
	} else {
		log.Printf("Container %s already exists, skipping creation.", containerName)
	}

	return "", nil
}

// StopContainer stops a container by its ID
func (r *RuntimeManager) StopContainer(containerID string) error {
	ctx := context.Background()
	timeout := 10
	StopOptions := container.StopOptions{
		Timeout: &timeout,
	}
	if err := r.DockerClient.Client.ContainerStop(ctx, containerID, StopOptions); err != nil {
		return fmt.Errorf("failed to stop container %s: %v", containerID, err)
	}
	return nil
}

// DeletePod deletes all containers of a pod including the pause container
func (r *RuntimeManager) DeletePod(podUUID string) error {
	ctx := context.Background()

	// Get the list of containers with the pod name prefix
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "pod_uid="+podUUID)

	containers, err := r.DockerClient.Client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers for pod %s: %v", podUUID, err)
	}

	// Stop and remove each container
	for _, cnt := range containers {
		if err := r.StopContainer(cnt.ID); err != nil {
			log.Printf("Failed to stop container %s: %v", cnt.ID, err)
		}

		if err := r.DockerClient.Client.ContainerRemove(ctx, cnt.ID, container.RemoveOptions{}); err != nil {
			log.Printf("Failed to remove container %s: %v", cnt.ID, err)
		}
		fmt.Printf("Deleted container %s\n", cnt.ID)
	}
	return nil
}
