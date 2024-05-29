package apiobject

type Pod struct {
	Name       string            `yaml:"name"`
	Status     string            `yaml:"status"`
	IP         string            `yaml:"ip"`
	Labels     map[string]string `yaml:"labels"`
	Containers []Container       `yaml:"containers"`
	NodeName   string            `yaml:"nodename"`
}

type Container struct {
	Name    string   `yaml:"name"`
	Image   string   `yaml:"image"`
	Command []string `yaml:"command"`
	Ports   []Port   `yaml:"ports"`
}
