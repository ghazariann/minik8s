package apiobject

// APIObject is the base structure embedded in all K8s-like objects.
type APIObject struct {
	APIVersion string   `yaml:"apiVersion" json:"apiVersion"`
	Kind       string   `yaml:"kind" json:"kind"`
	Metadata   Metadata `yaml:"metadata" json:"metadata"`
}

// Metadata contains the metadata for a Kubernetes object.
type Metadata struct {
	Name      string            `yaml:"name" json:"name"`
	Labels    map[string]string `yaml:"labels" json:"labels"`
	UUID      string            `json:"uuid" yaml:"uuid"`
	Namespace string            `json:"namespace" yaml:"namespace" default:"default"`
}
