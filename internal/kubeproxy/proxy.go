package kubeproxy

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var NginxPodYamlPath = "/root/minik8s/testdata/nginxPod.yaml"

type KubeProxy struct {
	iptableManager IptableManager
	dnsManager     dnsManager
	knownServices  map[string]apiobject.ServiceStore
	knowsDns       map[string]apiobject.DnsStore
	Hostname       string
}

func NewKubeProxy() (*KubeProxy, error) {
	iptableManager := &IptableManager{
		stragegy:       "roundrobin",
		serviceToPod:   make(map[string][]string),
		serviceToChain: make(map[string][]string),
		chainToRule:    make(map[string][]string),
	}
	dnsManager := &dnsManager{}
	iptableManager.Initialize_iptables()
	hostname, _ := os.Hostname()
	return &KubeProxy{
		iptableManager: *iptableManager,
		dnsManager:     *dnsManager,
		knownServices:  make(map[string]apiobject.ServiceStore),
		knowsDns:       make(map[string]apiobject.DnsStore),
		Hostname:       hostname,
	}, nil
}

// getAllDns
func (p *KubeProxy) GetAllDns() ([]apiobject.DnsStore, error) {
	url := configs.GetApiServerUrl() + configs.DnssUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list dns: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var dnss []apiobject.DnsStore
	if err := json.Unmarshal(body, &dnss); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	return dnss, nil
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
func UpdateDnsStatus(dns apiobject.DnsStore) error {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.DnsUrl+"?name=%s", dns.Metadata.Name)

	dnsJson, err := json.Marshal(dns)
	if err != nil {
		return fmt.Errorf("marshalling dns: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(dnsJson))
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

// AppendHostEntries appends multiple IP addresses for the same hostname to /etc/hosts
func appendHostEntries(ips []string, hostname string) error {
	// Open the /etc/hosts file in append mode
	file, err := os.OpenFile("/etc/hosts", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open /etc/hosts: %v", err)
	}
	defer file.Close()

	// Write each IP-hostname pair to the file
	for _, ip := range ips {
		entry := fmt.Sprintf("%s\t%s", ip, hostname)
		if _, err := file.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("could not write to /etc/hosts: %v", err)
		}
	}

	return nil
}

func deleteHostEntries(hostname string) error {
	// Read the /etc/hosts file
	file, err := os.Open("/etc/hosts")
	if err != nil {
		return fmt.Errorf("could not open /etc/hosts: %v", err)
	}
	defer file.Close()

	// Create a temporary file to store the updated hosts file
	tmpFile, err := os.CreateTemp("", "hosts")
	if err != nil {
		return fmt.Errorf("could not create temp file: %v", err)
	}
	defer tmpFile.Close()

	// Read through the /etc/hosts file and write to the temporary file
	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tmpFile)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, hostname) {
			if _, err := writer.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("could not write to temp file: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading /etc/hosts: %v", err)
	}

	writer.Flush()

	// Replace the original /etc/hosts file with the updated one
	if err := os.Rename(tmpFile.Name(), "/etc/hosts"); err != nil {
		return fmt.Errorf("could not replace /etc/hosts: %v", err)
	}

	return nil
}
func getAllIps(paths []apiobject.Path) string {
	var ipList []string
	for _, path := range paths {
		ipList = append(ipList, path.ServiceIp)
	}
	return strings.Join(ipList, " ")
}

func UpdateServiceStatus(service apiobject.ServiceStore) error {
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
func reloadNginx(containerName string) error {
	cmd := exec.Command("docker", "exec", containerName, "nginx", "-s", "reload")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to reload Nginx: %v, output: %s", err, string(output))
	}
	fmt.Printf("Nginx reload output: %s\n", string(output))
	return nil
}

func GetAllServiceIps() ([]string, error) {
	url := configs.GetApiServerUrl() + configs.DnsServiceIPUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Unmarshal the JSON response into a slice of PodStore
	serviceIp := string(body)
	// log.Printf("Response body: %s", body)
	// if err := json.Unmarshal(body, &serviceIp); err != nil {
	// 	log.Fatalf("Error unmarshalling response body: %v", err)
	// }
	var serviceIps = []string{serviceIp}
	return serviceIps, nil
}
func (p *KubeProxy) DnsRoutine() error {
	dnss, err := p.GetAllDns()
	if err != nil {
		return err
	}
	// detect deleted dns
	newDns := apiobject.DnsStore{}
	for dnsName, _ := range p.knowsDns {
		found := false
		for _, newDns = range dnss {
			if newDns.Metadata.Name == dnsName {
				found = true
				break
			}
		}
		if !found {
			fmt.Println("delete dns", dnsName)
			p.dnsManager.DeleteDns(newDns)
			delete(p.knowsDns, dnsName)
		}
	}
	// check if the phase is pending
	for _, dns := range dnss {
		if dns.Status.Phase == "pending" {
			fmt.Println("create dns", dns.Metadata.Name)
			// get all services ips in list fro dns.Spec.Paths
			serviceIps, _ := GetAllServiceIps()
			appendHostEntries(serviceIps, dns.Spec.Hostname)
			p.dnsManager.AddDns(dns)
			// add to /etc/hosts
			// ipList := getAllIps(dns.Spec.Paths)
			// appendHostEntry(ipList, dns.Spec.Hostname)
			dns.Status.Phase = "running"
			UpdateDnsStatus(dns)
			p.knowsDns[dns.Metadata.Name] = dns
			reloadNginx("dns-nginx-vahag-master_nginx") // TODO change to dynamic
		}
	}
	return nil
}

func (p *KubeProxy) ServiceRoutine() error {

	services, err := p.GetAllServices()
	if err != nil {
		return err
	}
	// Detect deleted services
	for serviceName, _ := range p.knownServices {
		found := false
		for _, newService := range services {
			if newService.Metadata.Name == serviceName {
				found = true
				break
			}
		}
		if !found {
			fmt.Println("delete service", serviceName)
			p.iptableManager.CleanIpTables(serviceName)
			delete(p.knownServices, serviceName)
		}
	}
	for _, service := range services {
		p.knownServices[service.Metadata.Name] = service
		if service.Status.Phase == "pending" {
			fmt.Println("create service", service.Metadata.Name)
			err := p.iptableManager.CreateService(service)
			if err != nil {
				log.Printf("KUBEPROXY: ServiceRoutine: CreateService failed: %v", err)
				continue
			}
			service.Status.Phase = "running"
			UpdateServiceStatus(service)
		}
	}

	return nil
}

func (proxy *KubeProxy) CreateNginxPod() {
	// get all pods
	url := configs.GetApiServerUrl() + configs.PodsURL
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list pods: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var pods []apiobject.Pod
	if err := json.Unmarshal(body, &pods); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	// loop and find same nod and same nginx label
	for _, pod := range pods {
		if pod.Spec.NodeName == proxy.Hostname && strings.Contains(pod.Metadata.Name, "dns-nginx") {
			return
		}
	}
	// create nginx pod
	data, err := os.ReadFile(NginxPodYamlPath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var pod apiobject.Pod
	if err := yaml.Unmarshal(data, &pod); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}
	pod.Metadata.Name = fmt.Sprintf("dns-nginx-%s", proxy.Hostname)
	pod.Spec.NodeName = proxy.Hostname

	jsonData, err := json.Marshal(pod)
	if err != nil {
		log.Fatalf("Error converting pod data to JSON: %v", err)
	}

	resp, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

}

func (p *KubeProxy) WatchService() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		err := p.ServiceRoutine()
		if err != nil {
			log.Printf("KUBEPROXY error: %v", err)
			continue
		}
		err = p.DnsRoutine()
		if err != nil {
			log.Printf("KUBEPROXY error: %v", err)
			continue
		}

	}
}
