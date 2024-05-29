package kubelet

import (
	"minik8s/internal/apiobject"
	"minik8s/internal/types"
)

func (r *RuntimeManager) GenerateContainerConfig(pod *apiobject.PodStore, container *apiobject.Container, pauseContainerID string) (*types.ContainerConfig, error) {

	containerLabels := make(map[string]string)
	for key, value := range pod.Metadata.Labels {
		containerLabels[key] = value
	}

	containerLabels["pod_name"] = pod.Metadata.Name
	containerLabels["pod_uid"] = string(pod.Metadata.UUID)
	containerLabels["is_pause"] = "false"
	containerLabels["pod_namespace"] = pod.Metadata.Namespace
	pauseRef := "container:" + pauseContainerID

	// // [Binds] Process the volumeMounts of the incoming configuration container and the volumeMounts of creating the container
	// // containerBinds, err := r.parseVolumeBinds(pod.Spec.Volumes, container.VolumeMounts)

	// if err != nil {
	// 	return nil, err
	// }

	config := types.ContainerConfig{
		Image:       container.Image,
		Entrypoint:  container.Command,
		Labels:      containerLabels,
		PidMode:     pauseRef,
		IpcMode:     pauseRef,
		NetworkMode: pauseRef,
		Volumes:     nil,
		// Binds:       containerBinds,
		// CPUResourceLimit: int64(container.Resources.Limits.CPU),
		// MemoryLimit:      int64(container.Resources.Limits.Memory),
	}
	return &config, nil
}
