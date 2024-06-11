package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"os"
	"time"

	"net/http"

	"gopkg.in/yaml.v2"
)

// TODO change to env var
var NginxServiceYamlPath = "/root/minik8s/testdata/nginxService.yaml"
var NginxDnsYamlPath = "/root/minik8s/testdata/nginxDns.yaml"

type DnsController interface {
	Run()
}

type dnsController struct {
	hostList         []string
	nginxServiceName string
	nginxServiceIp   string
}

func NewDnsController() DnsController {

	return &dnsController{
		hostList: make([]string, 0),
	}
}

func (dc *dnsController) DnsCreateHandler(dnsStore *apiobject.DnsStore) {

	if dnsStore.Spec.Hostname == "" {
		return
	}

	// nginxConfig := FormatNginxConfig(*dnsStore.ToDns())

	newHostEntry := dc.nginxServiceIp + " " + dnsStore.Spec.Hostname
	dc.hostList = append(dc.hostList, newHostEntry)

	// // 创建hostUpdate消息
	// hostUpdate := &entity.HostUpdate{
	// 	Action:    message.CREATE,
	// 	DnsTarget: dnsStore,
	// 	DnsConfig: nginxConfig,
	// 	HostList:  dc.hostList,
	// }

	// // TODO: 通知所有的节点进行hosts的修改
	// k8log.DebugLog("Dns-Controller", "DnsCreateHandler: publish hostUpdate")
	// message.PubelishUpdateHost(hostUpdate)
}

func (dc *dnsController) CreateNginxService() {
	data, err := os.ReadFile(NginxServiceYamlPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	var service apiobject.Service
	if err := yaml.Unmarshal(data, &service); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	jsonData, err := json.Marshal(service)
	if err != nil {
		log.Fatalf("Error converting service data to JSON: %v", err)
	}

	url := configs.GetApiServerUrl() + configs.ServicesUrl
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	dc.nginxServiceName = service.Metadata.Name

}

func (dc *dnsController) GetNginxServiceIP() {

	var nginxSvc apiobject.ServiceStore

	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.ServiceUrl+"?name=%s", dc.nginxServiceName)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// log.Printf("Response body: %s", body)
	err = json.Unmarshal(body, &nginxSvc)
	if err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}
	dc.nginxServiceIp = nginxSvc.Spec.ClusterIP
}

func (dc *dnsController) CreateNginxDns() {
	data, err := os.ReadFile(NginxDnsYamlPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}
	var dns apiobject.Dns
	if err := yaml.Unmarshal(data, &dns); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}
	jsonData, err := json.Marshal(dns)
	if err != nil {
		log.Fatalf("Error converting service data to JSON: %v", err)
	}
	url := configs.GetApiServerUrl() + configs.DnsUrl
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

}
func (dc *dnsController) UpdateServiceIp() error {
	// Prepare the IP as JSON
	url := configs.GetApiServerUrl() + configs.DnsServiceIPUrl
	ip := dc.nginxServiceIp
	jsonData := []byte(`"` + ip + `"`) // Encoded as a JSON string

	// Make the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to update IP: %s", resp.Status)
		return err
	}

	log.Println("Service IP updated successfully")
	return nil

}
func (dc *dnsController) Run() {
	// sleep for a while so apiserver will start
	time.Sleep(2 * time.Second)
	// dc.CreateNginxService()
	// dc.GetNginxServiceIP()
	// dc.UpdateServiceIp()
	// check periodically
}
