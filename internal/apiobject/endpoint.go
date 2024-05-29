package apiobject

type Endpoint struct {
	APIObject `json:",inline" yaml:",inline"`
	PodUUID   string   `yaml:"podUUID"`
	IP        string   `yaml:"ip"`
	Ports     []string `yaml:"port"`
}
