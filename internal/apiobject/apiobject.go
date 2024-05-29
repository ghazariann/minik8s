package apiobject

// APIObject is the base structure embedded in all K8s-like objects.
type APIObject struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
}

type Metadata struct {
	Name      string            `yaml:"name"`
	Labels    map[string]string `yaml:"labels"`
	UUID      string            `json:"uuid" yaml:"uuid"`
	Namespace string            `json:"namespace" yaml:"namespace" default:"default"`
}
