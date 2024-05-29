package apiobject

import (
	"time"
)

// Deployment represents a Kubernetes Deployment.
type Deployment struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      DeploymentSpec `yaml:"spec" json:"spec"`
}

// DeploymentSpec defines the specification of a Deployment.
type DeploymentSpec struct {
	Replicas int           `yaml:"replicas" json:"replicas"`
	Selector LabelSelector `yaml:"selector" json:"selector"`
	Template PodTemplate   `yaml:"template" json:"template"`
}

// DeploymentStore represents a stored Deployment with status.
type DeploymentStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      DeploymentSpec   `yaml:"spec" json:"spec"`
	Status    DeploymentStatus `yaml:"status" json:"status"`
}

// DeploymentStatus represents the status of a Deployment.
type DeploymentStatus struct {
	ReadyReplicas int `json:"readyReplicas" yaml:"readyReplicas"`
}

// DeploymentCondition represents the condition of a Deployment.
type DeploymentCondition struct {
	Type           string    `json:"type" yaml:"type"`
	Status         string    `json:"status" yaml:"status"`
	LastUpdateTime time.Time `json:"lastUpdateTime" yaml:"lastUpdateTime"`
	Reason         string    `json:"reason" yaml:"reason"`
	Message        string    `json:"message" yaml:"message"`
}

// ToDeploymentStore converts a Deployment to DeploymentStore.
func (d *Deployment) ToDeploymentStore() *DeploymentStore {
	return &DeploymentStore{
		APIObject: d.APIObject,
		Spec:      d.Spec,
		Status:    DeploymentStatus{}, // Initially empty
	}
}

// ToDeployment converts a DeploymentStore back to Deployment.
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
