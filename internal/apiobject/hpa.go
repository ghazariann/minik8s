package apiobject

type HpaMetrics struct {
	CPUPercent float64 `yaml:"cpuPercent" json:"cpuPercent"`
	MemPercent float64 `yaml:"memPercent" json:"memPercent"`
}

type HpaSpec struct {
	MinReplicas    int           `yaml:"minReplicas" json:"minReplicas"`
	MaxReplicas    int           `yaml:"maxReplicas" json:"maxReplicas"`
	ScaleTargetRef APIObject     `yaml:"scaleTargetRef" json:"wscaleTargetReforkload"`
	Interval       int           `yaml:"interval" json:"interval"` // in seconds
	Selector       LabelSelector `yaml:"selector" json:"selector"`
	Metrics        HpaMetrics    `yaml:"metrics" json:"metrics"`
}

type Hpa struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      HpaSpec `yaml:"spec" json:"spec"`
}

type HpaStore struct {
	APIObject `yaml:",inline" json:",inline"`
	Spec      HpaSpec   `yaml:"spec" json:"spec"`
	Status    HpaStatus `yaml:"status" json:"status"`
}

type HpaStatus struct {
	CurrentReplicas   int     `yaml:"currentReplicas" json:"currentReplicas"`
	CurrentCPUPercent float64 `yaml:"currentCPUPercent" json:"currentCPUPercent"`
	CurrentMemPercent float64 `yaml:"currentMemPercent" json:"currentMemPercent"`
}

// ToStore
func (hpa *Hpa) ToStore() *HpaStore {
	return &HpaStore{
		APIObject: hpa.APIObject,
		Spec:      hpa.Spec,
		Status:    HpaStatus{},
	}
}
