package apiobject

type Path struct {
	UrlPath     string `json:"urlPath" yaml:"urlPath"`
	ServiceName string `json:"serviceName" yaml:"serviceName"`
	ServicePort string `json:"servicePort" yaml:"servicePort"`
	ServiceIp   string `json:"serviceIp" yaml:"serviceIp"`
}

type DnsSpec struct {
	Hostname string `json:"hostname" yaml:"hostname"`
	Paths    []Path `json:"paths" yaml:"paths"`
}

type Dns struct {
	APIObject `json:",inline" yaml:",inline"`
	Spec      DnsSpec `json:"spec" yaml:"spec"`
}

type DnsStatus struct {
	Phase string `json:"phase" yaml:"phase"`
	// LastUpdated time.Time `yaml:"lastUpdated" json:"lastUpdated"`
}

type DnsStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      DnsSpec   `yaml:"spec" json:"spec"`
	Status    DnsStatus `yaml:"status" json:"status"`
}

func (d *Dns) ToDnsStore() *DnsStore {
	return &DnsStore{
		APIObject: d.APIObject,
		Spec:      d.Spec,
	}
}

func (ds *DnsStore) ToDns() *Dns {
	return &Dns{
		APIObject: ds.APIObject,
		Spec:      ds.Spec,
	}
}
