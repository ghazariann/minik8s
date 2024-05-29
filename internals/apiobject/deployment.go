package apiobject

type Deployment struct {
	APIObject `yaml:",inline"`
	Spec      DeploymentSpec `yaml:"spec"`
}

type DeploymentSpec struct {
	Replicas int           `yaml:"replicas"`
	Selector LabelSelector `yaml:"selector"`
	Template PodTemplate   `yaml:"template"`
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
