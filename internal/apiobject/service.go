package apiobject

// type Service struct {
// 	Name       string            `yaml:"name"`
// 	Selector   map[string]string `yaml:"selector"`
// 	Port       int               `yaml:"port"`
// 	TargetPort int               `yaml:"targetPort"`
// 	IP         string            `yaml:"ip"`
// }

// // 创建简单的负载均衡策略，轮询（Round Robin）策略
// var nextPodIndex = 0

// func (s *Service) Forward() *Pod {
// 	pods := getPodsBySelector(s.Selector)
// 	if len(pods) == 0 {
// 		return nil
// 	}
// 	pod := pods[nextPodIndex%len(pods)]
// 	nextPodIndex++
// 	return pod
// }

// // 根据Selector获取Pods列表
//
//	func getPodsBySelector(selector map[string]string) []*Pod {
//		var selectedPods []*Pod
//		// 这里需要添加逻辑来筛选符合Selector的Pods
//		return selectedPods
//	}

type ServiceStatus struct {
	Endpoints []Endpoint
	Phase     string
}

type ServicePort struct {
	Port       int    `yaml:"port"`
	TargetPort int    `yaml:"targetPort"`
	Protocol   string `yaml:"protocol"`
}

type ServiceSpec struct {
	Selector map[string]string `yaml:"selector"`
	Ports    []ServicePort     `yaml:"ports"`
	// Type      string            `yaml:"type"`
	// ClusterIP string            `yaml:"clusterIP"`
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
