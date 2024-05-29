package apiobject

import (
	"time"
)

type Deployment struct {
	APIObject `yaml:",inline"`
	Spec      DeploymentSpec `yaml:"spec"`
}

type DeploymentSpec struct {
	Replicas int           `yaml:"replicas"`
	Selector LabelSelector `yaml:"selector"`
	Template PodTemplate   `yaml:"template"`
}

type DeploymentStore struct {
	APIObject `json:",inline" yaml:",inline"`
	Spec      DeploymentSpec   `json:"spec" yaml:"spec"`
	Status    DeploymentStatus `json:"status" yaml:"status"`
}

type DeploymentStatus struct {
	Replicas      int                   `json:"replicas" yaml:"replicas"`
	ReadyReplicas int                   `json:"readyReplicas" yaml:"readyReplicas"`
	Conditions    []DeploymentCondition `json:"conditions" yaml:"conditions"`
}

type DeploymentCondition struct {
	Type           string    `json:"type" yaml:"type"`
	Status         string    `json:"status" yaml:"status"`
	LastUpdateTime time.Time `json:"lastUpdateTime" yaml:"lastUpdateTime"`
	Reason         string    `json:"reason" yaml:"reason"`
	Message        string    `json:"message" yaml:"message"`
}

func (d *Deployment) ToDeploymentStore() *DeploymentStore {
	return &DeploymentStore{
		APIObject: d.APIObject,
		Spec:      d.Spec,
		Status:    DeploymentStatus{}, // Initially empty
	}
}

func (ds *DeploymentStore) ToDeployment() *Deployment {
	return &Deployment{
		APIObject: ds.APIObject,
		Spec:      ds.Spec,
	}
}

type LabelSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}

type PodTemplate struct {
	Metadata Metadata `yaml:"metadata"`
	Spec     PodSpec  `yaml:"spec"`
}

type PodSpec struct {
	Containers []Container `yaml:"containers"`
}

type Port struct {
	ContainerPort int `yaml:"containerPort"`
}
