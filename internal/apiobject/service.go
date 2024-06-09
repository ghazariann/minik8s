package apiobject

type ServiceStatus struct {
	Endpoints []Endpoint `json:"endpoints" yaml:"endpoints"`
	Phase     string     // active, pending, failed
}

type ServicePort struct {
	Port       int    `yaml:"port" json:"port"`
	TargetPort int    `yaml:"targetPort" json:"targetPort"`
	Protocol   string `yaml:"protocol" json:"protocol"`
	NodePort   int    `yaml:"nodePort"`
}

type ServiceSpec struct {
	Selector  map[string]string `yaml:"selector" json:"selector"`
	Ports     []ServicePort     `yaml:"ports" json:"ports"`
	Type      string            `yaml:"type" json:"type"`
	ClusterIP string            `yaml:"clusterIP" json:"clusterIP"`
}

type Service struct {
	APIObject `json:",inline" yaml:",inline"`
	Spec      ServiceSpec `json:"spec" yaml:"spec"`
}

type ServiceStore struct {
	APIObject `json:",inline" yaml:",inline"`
	Spec      ServiceSpec   `json:"spec" yaml:"spec"`
	Status    ServiceStatus `json:"status" yaml:"status"`
}

func (s *Service) ToServiceStore() *ServiceStore {
	return &ServiceStore{
		APIObject: s.APIObject,
		Spec:      s.Spec,
	}
}
