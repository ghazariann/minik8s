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
func (p *KubeProxy) GetAllServices() ([]apiobject.ServiceStore, error) {
	url := configs.GetApiServerUrl() + configs.ServicesUrl
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	var services []apiobject.ServiceStore

	if err := json.Unmarshal(body, &services); err != nil {
		return nil, fmt.Errorf("unmarshalling response body: %v", err)
	}
	return services, nil
}
func (p *KubeProxy) UpdateServiceStatus(service apiobject.ServiceStore) error {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.ServiceStoreURL+"?name=%s", service.Metadata.Name)

	serviceJson, err := json.Marshal(service)
	if err != nil {
		return fmt.Errorf("marshalling service: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(serviceJson))
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %v", err)
	}
	defer resp.Body.Close()
	return nil
}
func (p *KubeProxy) ServiceRoutine() error {

	services, err := p.GetAllServices()
	if err != nil {
		return err
	}
	for _, service := range services {
		if service.Status.Phase == "pending" {
			fmt.Println("create service", service.Metadata.Name)
			p.iptableManager.CreateService(service)
			service.Status.Phase = "running"
			p.UpdateServiceStatus(service)
		}
	}
	return nil
}

func (p *KubeProxy) WatchService() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		err := p.ServiceRoutine()
		if err != nil {
			log.Printf("KUBEPROXY error: %v", err)
			continue
		}
	}
}
