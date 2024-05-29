package kubelet

import (
	"context"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerClient struct {
	Client *client.Client
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
	reader, err := d.Client.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(io.Discard, reader)
	return err
}

// ImageExists checks if a Docker image is already pulled
func (d *DockerClient) ImageExists(imageName string) bool {
	images, err := d.Client.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		log.Printf("Error listing images: %v", err)
		return false
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageName {
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
