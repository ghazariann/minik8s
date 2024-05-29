package kubelet

import (
	"context"
	"fmt"
	"log"
	"minik8s/internal/apiobject" // Ensure correct import path

	weave "minik8s/utils"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
)

type RuntimeManager struct {
	DockerClient *DockerClient
}

// NewRuntimeManager initializes and returns a new RuntimeManager
func NewRuntimeManager(dockerClient *DockerClient) *RuntimeManager {
	return &RuntimeManager{DockerClient: dockerClient}
}

// CreatePod creates a pod with the specified configuration
func (r *RuntimeManager) CreatePod(pod *apiobject.PodStore) error {
	ctx := context.Background()
	pauseID := pod.Metadata.Name + "_pause"
	// Pull the pause image
	pauseImage := "registry.aliyuncs.com/google_containers/pause:3.6"
	if !r.DockerClient.ImageExists(pauseImage) {
		if err := r.DockerClient.PullImage(pauseImage); err != nil {
			log.Printf("Failed to pull pause image %s: %v", pauseImage, err)
			return err
		}
	}
	// Create the pause container if not present
	pauseContainerName := pod.Metadata.Name + "_pause"
	if !r.DockerClient.ContainerExists(pauseContainerName) {
		pauseCntConfig := container.Config{
			Image:   pauseImage,
			Volumes: nil,
			Env:     nil,
			//TODO
		}
		pauseHostConfig := container.HostConfig{IpcMode: "shareable"}
		pauseResp, err := r.DockerClient.Client.ContainerCreate(ctx, &pauseCntConfig, &pauseHostConfig, nil, nil, pauseContainerName)
		pauseID = pauseResp.ID
		if err != nil {
			log.Printf("Failed to create pause container: %v", err)
			return err
		}

		// Start the pause container
		if err := r.DockerClient.Client.ContainerStart(ctx, pauseID, container.StartOptions{}); err != nil {
			log.Printf("Failed to start pause container: %v", err)
			return err
		}
		// [Weave网络] 为pause容器添加网络
		if pod.Status.PodIP == "" {
			res, err := weave.AttachContainer(pauseID)
			if err != nil {
				log.Fatal("Pause Container", err.Error()+res)
				return err
			}

			// TODO: add podIp to pod status
			pod.Status.PodIP = res
			log.Printf("WeaveAttach", "WeaveAttach res "+res)
		}
	}

	// Iterate over each container in the pod
	for _, containerSpec := range pod.Spec.Containers {
		// Check and pull the container image if not present
		if !r.DockerClient.ImageExists(containerSpec.Image) {
			if err := r.DockerClient.PullImage(containerSpec.Image); err != nil {
				log.Printf("Failed to pull image %s: %v", containerSpec.Image, err)
				continue
			}
		}
		pauseRef := "container:" + pauseID
		contConf := container.Config{
			Image: containerSpec.Image,
			Cmd:   containerSpec.Command,
			// Env:          option.Env,
			// Tty:          option.Tty,
			// Labels:       option.Labels,
			// Entrypoint:   option.Entrypoint,
			// Volumes:      option.Volumes,
			// ExposedPorts: exposedPortSet,
		}
		hostConf := container.HostConfig{
			// Binds:        option.Binds,
			// PortBindings: option.PortBindings,

			// VolumesFrom:  option.VolumesFrom,
			// Links:        option.Links,
			// Resources: container.Resources{
			// 	Memory:   option.MemoryLimit,
			// 	NanoCPUs: option.CPUResourceLimit,
			// },
			PidMode:     container.PidMode(pauseRef),
			IpcMode:     container.IpcMode(pauseRef),
			NetworkMode: container.NetworkMode(pauseRef),
		}
		// Create the application container if not present
		containerName := fmt.Sprintf("%s_%s", pod.Metadata.Name, containerSpec.Name)
		if !r.DockerClient.ContainerExists(containerName) {
			resp, err := r.DockerClient.Client.ContainerCreate(ctx, &contConf, &hostConf, &network.NetworkingConfig{}, nil, containerName)
			if err != nil {
				log.Printf("Failed to create container %s: %v", containerSpec.Image, err)
				continue
			}

			// Start the application container
			if err := r.DockerClient.Client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
				log.Printf("Failed to start container %s: %v", resp.ID, err)
				continue
			}
			fmt.Printf("Started container %s with ID %s\n", containerSpec.Image, resp.ID)
		} else {
			log.Printf("Container %s already exists, skipping creation.", containerName)
		}
	}

	return nil
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
func (r *RuntimeManager) DeletePod(podName string) error {
	ctx := context.Background()

	// Get the list of containers with the pod name prefix
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", podName)

	containers, err := r.DockerClient.Client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return fmt.Errorf("failed to list containers for pod %s: %v", podName, err)
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
