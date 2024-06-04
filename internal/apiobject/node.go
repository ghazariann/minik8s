package apiobject

import "time"

type Node struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      NodeSpec `yaml:"spec" json:"spec"`
}

type NodeStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      NodeSpec   `yaml:"spec" json:"spec"`
	Status    NodeStatus `yaml:"status" json:"status"`
}
type NodeSpec struct {
	IP string `json:"ip" yaml:"ip"`
}

type NodeStatus struct {
	Condition  string    `json:"condition" yaml:"condition"` // ready, unknown
	CpuPercent float64   `json:"cpuPercent" yaml:"cpuPercent"`
	MemPercent float64   `json:"memPercent" yaml:"memPercent"`
	NumPods    int       `json:"numPods" yaml:"numPods"`
	UpdateTime time.Time `json:"updateTime" yaml:"updateTime"`
}

// ToNodeStore converts a Node to NodeStore.
func (d *Node) ToStore() *NodeStore {
	return &NodeStore{
		APIObject: d.APIObject,
		Spec:      d.Spec,
		Status:    NodeStatus{}, // Initially empty
	}
}
