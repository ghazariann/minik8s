package types

import "github.com/docker/go-connections/nat"

type ContainerConfig struct {
	Entrypoint   []string
	Image        string
	Volumes      map[string]struct{}
	Labels       map[string]string
	ExposedPorts map[string]struct{}

	// HostConfig
	VolumesFrom  []string
	Links        []string
	NetworkMode  string
	PidMode      string      // [PidMode] PID namespace to use for the container
	IpcMode      string      // [IPC Mode ]IPC namespace to use for the container
	Binds        []string    // List of volume bindings for this container
	PortBindings nat.PortMap // List of port bindings for this container

	CPUResourceLimit int64
	MemoryLimit      int64
}
