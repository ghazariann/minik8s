package apiobject

import "time"

const (
	// PodPending代表Pod处于Pending状态
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

// type ContainerPort struct {
// 	Name          string `yaml:"name" json:"name"`
// 	HostPort      string `yaml:"hostPort" json:"hostPort"`
// 	ContainerPort string `yaml:"containerPort" json:"containerPort"`
// 	Protocol      string `yaml:"protocol" json:"protocol"`
// 	HostIP        string `yaml:"hostIP" json:"hostIP"`
// }

// Container represents a container within a Pod.
type Container struct {
	Name    string   `yaml:"name" json:"name"`
	Image   string   `yaml:"image" json:"image"`
	Command []string `yaml:"command" json:"command"`
	Ports   []Port   `yaml:"ports" json:"ports"`
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
