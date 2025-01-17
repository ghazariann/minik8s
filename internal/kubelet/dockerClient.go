package kubelet

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	Client *client.Client
}

func CpuToMillicores(cpu string) (int64, error) {
	cpu = strings.TrimSpace(cpu)
	if strings.HasSuffix(cpu, "m") {
		// Remove 'm' suffix and parse as millicores
		number, err := strconv.ParseInt(cpu[:len(cpu)-1], 10, 64)
		if err != nil {
			return 0, err
		}
		return number, nil
	}

	// Assume it's a whole number representing full cores, convert to millicores
	number, err := strconv.ParseInt(cpu, 10, 64)
	if err != nil {
		return 0, err
	}
	return number * 1000, nil
}

func MemoryToBytes(mem string) (int64, error) {
	// Remove any whitespace
	mem = strings.TrimSpace(mem)

	// Get the number and unit
	lastChar := mem[len(mem)-2:]
	multiplier := int64(1) // Default to bytes

	switch lastChar {
	case "Mi":
		multiplier = 1024 * 1024
	case "Gi":
		multiplier = 1024 * 1024 * 1024
	case "Ki":
		multiplier = 1024
	default:
		// Handle simple byte case or any other non-standard case
		return strconv.ParseInt(mem, 10, 64)
	}

	// Extract the numeric part
	number, err := strconv.ParseInt(mem[:len(mem)-2], 10, 64)
	if err != nil {
		return 0, err
	}

	return number * multiplier, nil
}

// NewDockerClient initializes and returns a new Docker client
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &DockerClient{Client: cli}, nil
}

// PullImage pulls a Docker image
func (d *DockerClient) PullImage(imageName string) error {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(os.Stdout, reader)
	return err
}

// ImageExists checks if a Docker image is already pulled
func (d *DockerClient) ImageExists(images []image.Summary, imageName string) bool {
	// images, err := d.Client.ImageList(context.Background(), image.ListOptions{})
	// if err != nil {
	// 	log.Printf("Error listing images: %v", err)
	// 	return false
	// }
	for _, image := range images {
		for _, tag := range image.RepoTags {
			// fmt.Printf("tag: %s\n", tag)
			if tag == imageName || tag == imageName+":latest" {
				return true
			}
		}
	}
	return false
}

// ContainerExists checks if a Docker container is already created
func (d *DockerClient) ContainerExists(containerName string) bool {
	containers, err := d.Client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return false
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+containerName {
				return true
			}
		}
	}
	return false
}
func (d *DockerClient) GetContainerIDByName(containerName string) (string, error) {
	ctx := context.Background()
	containers, err := d.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, container := range containers {
		if containerName == strings.TrimPrefix(container.Names[0], "/") {
			return container.ID, nil
		}
	}

	return "", fmt.Errorf("container %s not found", containerName)
}

// ListContainers lists all Docker containers
func (d *DockerClient) ListContainers() ([]types.Container, error) {
	ctx := context.Background()
	containers, err := d.Client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// ListPodContainers lists all containers associated with pods and returns a map of podUUID to containerName
func (d *DockerClient) ListPodContainers() (map[string]string, error) {
	ctx := context.Background()

	// Create filter arguments to filter containers by the "pod_uid" label
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", "pod_uid")

	// List containers with the specified label
	containers, err := d.Client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, err
	}

	// Initialize a map to store pod UUIDs to container names
	podContainers := make(map[string]string)

	// Iterate through the containers and populate the map
	for _, cnt := range containers {
		podUID := cnt.Labels["pod_uid"]
		containerID := cnt.ID
		// containerName := cnt.Names[0] // Assuming the container name is in the format "/container_name"
		podContainers[containerID] = podUID
	}

	return podContainers, nil
}

func GetIDFromContainerName(containerName string) string {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Error creating docker client: %v", err)
		return ""
	}

	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		log.Printf("Error listing containers: %v", err)
		return ""
	}

	for _, container := range containers {
		if containerName == strings.TrimPrefix(container.Names[0], "/") {
			return container.ID
		}
	}

	return ""
}
