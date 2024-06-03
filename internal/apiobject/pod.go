package apiobject

import "time"

const (
	PodPending     = "Pending"
	PodRunning     = "Running"
	PodSucceeded   = "Succeeded"
	PodFailed      = "Failed"
	PodUnknown     = "Unknown"
	PodTerminating = "Terminating"
)

// Pod represents a Kubernetes Pod.
type Pod struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      PodSpec `yaml:"spec" json:"spec"`
}

// PodSpec defines the specification of a Pod.
type PodSpec struct {
	Containers []Container `yaml:"containers" json:"containers"`
	NodeName   string      `yaml:"nodeName" json:"nodeName"`
}

// PodStatus represents the status of a Pod.
type PodStatus struct {
	PodIP      string    `yaml:"podIP" json:"podIP"`
	Phase      string    `yaml:"phase" json:"phase"`
	UpdateTime time.Time `yaml:"lastUpdateTime" json:"lastUpdateTime"`
	CpuPercent float64   `yaml:"cpuPercent" json:"cpuPercent"`
	MemPercent float64   `yaml:"memPercent" json:"memPercent"`
}

// PodStore represents a stored Pod with status.
type PodStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      PodSpec   `yaml:"spec" json:"spec"`
	Status    PodStatus `yaml:"status" json:"status"`
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

// Container represents a container within a Pod.
type Container struct {
	Name      string             `yaml:"name" json:"name"`
	Image     string             `yaml:"image" json:"image"`
	Command   []string           `yaml:"command" json:"command"`
	Args      []string           `yaml:"args" json:"args"`
	Ports     []Port             `yaml:"ports" json:"ports"`
	Resources ContainerResources `yaml:"resources"`
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
