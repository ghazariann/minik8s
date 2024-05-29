package kubeproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"
	"time"
)

type KubeProxy struct {
	iptableManager IptableManager
}

func NewKubeProxy() (*KubeProxy, error) {
	iptableManager := &IptableManager{
		stragegy:       "random",
		serviceToPod:   make(map[string][]string),
		serviceToChain: make(map[string][]string),
		chainToRule:    make(map[string][]string),
	}

	iptableManager.Initialize_iptables()
	return &KubeProxy{
		iptableManager: *iptableManager,
	}, nil
}
func (p *KubeProxy) GetAllServices() []apiobject.ServiceStore {
	url := configs.GetApiServerUrl() + configs.ServicesURL
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	var services []apiobject.ServiceStore

	if err := json.Unmarshal(body, &services); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	return services
}
func (p *KubeProxy) UpdateServiceStatus(service apiobject.ServiceStore) {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.ServiceStoreURL+"?name=%s", service.Metadata.Name)

	serviceJson, err := json.Marshal(service)
	if err != nil {
		log.Fatalf("Error encoding service data: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(serviceJson))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error updating service status: %v", err)
	}
	defer resp.Body.Close()
}
func (p *KubeProxy) ServiceRoutine() {

	services := p.GetAllServices()
	for _, service := range services {
		if service.Status.Phase == "pending" {
			fmt.Println("create service", service.Metadata.Name)
			p.iptableManager.CreateService(service)
			service.Status.Phase = "running"
			p.UpdateServiceStatus(service)
		}
	}
}

func (p *KubeProxy) WatchService() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		p.ServiceRoutine()
	}
}
