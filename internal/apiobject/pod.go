package apiobject

import (
	"time"

	"github.com/docker/docker/api/types"
)

const (
	PodPending     = "pending"
	PodRunning     = "running"
	PodSucceeded   = "Succeeded"
	PodFailed      = "Failed"
	PodUnknown     = "Unknown"
	PodTerminating = "Terminating"
)

// https://kubernetes.io/docs/concepts/storage/volumes/
type HostPath struct {
	Path string `json:"path" yaml:"path"`
	Type string `json:"type" yaml:"type"`
}

type Volume struct {
	Name     string   `json:"name" yaml:"name"`
	HostPath HostPath `json:"hostPath" yaml:"hostPath"`
}

// Pod represents a Kubernetes Pod.
type Pod struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      PodSpec `yaml:"spec" json:"spec"`
}

// PodStore represents a stored Pod with status.
type PodStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      PodSpec   `yaml:"spec" json:"spec"`
	Status    PodStatus `yaml:"status" json:"status"`
}

// PodSpec defines the specification of a Pod.
type PodSpec struct {
	Containers []Container `yaml:"containers" json:"containers"`
	NodeName   string      `yaml:"nodeName" json:"nodeName"`
	Volumes    []Volume    `json:"volumes" yaml:"volumes"`
}

// PodStatus represents the status of a Pod.
type PodStatus struct {
	PodIP             string                 `yaml:"podIP" json:"podIP"`
	Phase             string                 `yaml:"phase" json:"phase"`
	LastUpdated       time.Time              `yaml:"lastUpdateTime" json:"lastUpdateTime"`
	ContainerStatuses []types.ContainerState `json:"containerStatuses" yaml:"containerStatuses"`
	ContainerIDs      []string               `json:"containerIDs" yaml:"containerIDs"`
	CpuPercent        float64                `yaml:"cpuPercent" json:"cpuPercent"`
	MemPercent        float64                `yaml:"memPercent" json:"memPercent"`
}

type Port struct {
	ContainerPort int `yaml:"containerPort"`
}
type ResourceType struct {
	Cpu    string `yaml:"cpu" json:"cpu"`       // 256Mi
	Memory string `yaml:"memory" json:"memory"` // 256Mi
}

type ContainerResources struct {
	Limits   ResourceType `yaml:"limits"`
	Requests ResourceType `yaml:"requests"`
}

type EnvVar struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
}
type VolumeMount struct {
	Name      string `yaml:"name" json:"name"`
	MountPath string `yaml:"mountPath" json:"mountPath"`
}

// Container represents a container within a Pod.
type Container struct {
	Name         string             `yaml:"name" json:"name"`
	Image        string             `yaml:"image" json:"image"`
	Command      []string           `yaml:"command" json:"command"`
	Args         []string           `yaml:"args" json:"args"`
	Ports        []Port             `yaml:"ports" json:"ports"`
	Resources    ContainerResources `yaml:"resources"`
	Env          []EnvVar           `yaml:"env"`
	VolumeMounts []VolumeMount      `yaml:"volumeMounts" json:"volumeMounts"`
}

func (p *Pod) ToStore() *PodStore {
	return &PodStore{
		APIObject: p.APIObject,
		Spec:      p.Spec,
		Status:    PodStatus{},
	}
}

func (p *PodStore) ToPod() *Pod {
	return &Pod{
		APIObject: p.APIObject,
		Spec:      p.Spec,
	}
}
