package kubelet

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net"
	"net/http"
	"os"
	"time"
)

type Kubelet struct {
	Name            string
	RuntimeManager  *RuntimeManager
	knownPods       map[string]apiobject.PodStore
	knownContainers map[string]string
}

func IsRegisterd() bool {
	return false
}
func UnRegisterNode() {
	// if !IsRegisterd() {
	// 	log.Printf("Node not registered")
	// 	return
	// }
	log.Printf("Unregistering node")
	nodeIP, err := GetPrimaryIPv4Address()
	if err != nil {
		log.Printf("Error getting node IP: %v", err)
		return
	}

	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.NodesUrl+"?ip=%s", nodeIP)
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error unregistering node: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to unregister node: %s", string(body))
		return
	}

	log.Printf("Successfully unregistered node")
}

// NewKubelet initializes and returns a new Kubelet
func NewKubelet() (*Kubelet, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	runtimeManager := NewRuntimeManager(dockerClient)
	hostname, _ := os.Hostname()
	RegisterNode(hostname)

	return &Kubelet{
		RuntimeManager:  runtimeManager,
		knownPods:       map[string]apiobject.PodStore{},
		Name:            hostname,
		knownContainers: map[string]string{},
	}, nil
}
func (k *Kubelet) GetAllPods() ([]apiobject.PodStore, error) {
	url := configs.GetApiServerUrl() + configs.PodsURL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	var pods []apiobject.PodStore
	if err := json.Unmarshal(body, &pods); err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %v", err)
	}

	return pods, nil
}

//	func (k *Kubelet) routine() {
//		pods, _ := k.GetAllPods()
//		fmt.Printf("Kubelet %s: %d pods\n", k.Name, len(pods))
//	}
func (k *Kubelet) SyncContainers(knownContainers map[string]string, newContainers map[string]string, pods map[string]apiobject.PodStore) {
	// Check for missing containers
	for containerID, podUID := range knownContainers {
		if _, ok := newContainers[containerID]; !ok { // one or two does not matter
			// Container is missing, look for a matching pod
			if pod, podExists := pods[podUID]; podExists {
				// Pod exists, recreate the missing container
				log.Printf("Recreating missing container with id: %s", containerID)
				// list and find the container in pod.spec.containers
				k.RuntimeManager.CreatePod(&pod)
				if err := UpdatePodStatus(&pod); err != nil {
					log.Printf("Error updating pod status for %s: %v", pod.Metadata.Name, err)
				} else {
					log.Printf("Successfully updated pod status for %s to Running", pod.Metadata.Name)
					UpdateNodeStatus(&pod, "create")
				}

			}
		} else {
			if containerID == "" {
				continue
			}
			// check is container is running
			info, _ := k.RuntimeManager.GetInspectInfo(containerID)
			if info == nil {
				log.Printf("Container %s not found", containerID)
				continue
			}
			containerStatus := k.RuntimeManager.GetContainerState(info)
			if containerStatus.Status != "running" {
				// Container is not running, restart it
				log.Printf("Container %s is not running, restarting...", containerID)
				k.RuntimeManager.RestartContainer(containerID)
			}
		}

	}
}
func filterPodsByNodeName(pods []apiobject.PodStore, nodeName string) []apiobject.PodStore {
	filteredPods := []apiobject.PodStore{}
	for _, pod := range pods {
		if pod.Spec.NodeName == nodeName {
			filteredPods = append(filteredPods, pod)
		}
	}
	return filteredPods
}
func (k *Kubelet) MonitorAndManagePods() error {

	// Fetch all pods
	containers, _ := k.RuntimeManager.DockerClient.ListPodContainers()
	k.SyncContainers(k.knownContainers, containers, k.knownPods)
	k.knownContainers = containers

	pods, err := k.GetAllPods()
	// filter pods by node name

	pods = filterPodsByNodeName(pods, k.Name)

	if err != nil {
		return err
	}
	// Convert fetched pods to a map for easy comparison
	currentPods := map[string]apiobject.PodStore{}
	for _, pod := range pods {
		currentPods[pod.Metadata.UUID] = pod
	}

	// Detect deleted pods
	for podName := range k.knownPods {
		if _, exists := currentPods[podName]; !exists {
			log.Printf("Pod %s has been deleted. Cleaning up resources...", podName)
			if err := k.CleanUpPod(podName); err != nil {
				log.Printf("Error cleaning up pod %s: %v", podName, err)
			} else {
				log.Printf("Successfully cleaned up resources for pod %s", podName)
			}
		}
	}

	// Update knownPods to current state
	k.knownPods = currentPods

	// Process pending pods
	for _, pod := range pods {
		if pod.Status.Phase != apiobject.PodRunning {
			log.Printf("Pod %s is pending. Attempting to create containers...", pod.Metadata.Name)
			k.RuntimeManager.CreatePod(&pod)
			if err := UpdatePodStatus(&pod); err != nil {
				log.Printf("Error updating pod status for %s: %v", pod.Metadata.Name, err)
			} else {
				log.Printf("Successfully updated pod status to %s", pod.Status.Phase)
				UpdateNodeStatus(&pod, "create")
			}
		}
	}
	return nil
}

// CleanUpPod stops and removes all containers associated with the pod
func (k *Kubelet) CleanUpPod(podName string) error {
	log.Printf("Cleaning up resources for pod %s", podName)
	pod := k.knownPods[podName]
	k.RuntimeManager.DeletePod(podName)
	UpdateNodeStatus(&pod, "delete")
	return nil
}

// UpdatePodStatus sends a request to the API server to update the pod status
func UpdatePodStatus(pod *apiobject.PodStore) error {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.PodStoreUrl+"?name=%s", pod.Metadata.Name)
	podJson, err := json.Marshal(pod)
	if err != nil {
		return fmt.Errorf("failed to marshal pod status: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(podJson))
	if err != nil {
		return fmt.Errorf("failed to send update request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update pod status: %s", string(body))
	}

	return nil
}
func UpdateNodeStatus(pod *apiobject.PodStore, action string) error {
	GetUrl := fmt.Sprintf(configs.GetApiServerUrl()+configs.NodeUrl+"?name=%s", pod.Spec.NodeName)

	resp, err := http.Get(GetUrl)
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}
	defer resp.Body.Close()

	node := apiobject.NodeStore{}
	if err := json.NewDecoder(resp.Body).Decode(&node); err != nil {
		return fmt.Errorf("failed to decode node status: %v", err)
	}
	node.Status.UpdateTime = time.Now()
	if action == "delete" {
		node.Status.NumPods -= 1
		node.Status.MemPercent -= pod.Status.MemPercent
		node.Status.CpuPercent -= pod.Status.CpuPercent
		if node.Status.NumPods == 0 {
			node.Status.Condition = "idle"
		}
	} else if action == "create" {

		node.Status.NumPods += 1
		node.Status.MemPercent += pod.Status.MemPercent
		node.Status.CpuPercent += pod.Status.CpuPercent
		node.Status.Condition = "running"
	} else {
	}
	// else {
	// node.Status.MemPercent += pod.Status.MemPercent
	// node.Status.CpuPercent += pod.Status.CpuPercent
	// }
	Posturl := fmt.Sprintf(configs.GetApiServerUrl()+configs.NodeUrl+"?name=%s", pod.Spec.NodeName)

	jsonData, _ := json.Marshal(node)
	log.Printf("Node status: %s", jsonData)
	_, err = http.Post(Posturl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to marshal node status: %v", err)
	}
	return nil
}

func GetPrimaryIPv4Address() (string, error) {
	desiredInterfaceNames := []string{"ens3", "ens33", "eth0"}
	for _, name := range desiredInterfaceNames {
		iface, err := net.InterfaceByName(name)
		if err != nil {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("no interface found with the specified names")
}

// func CheckIfRegisterd() bool {
// 	url := configs.GetApiServerUrl() + configs.NodesUrl

// }
func RegisterNode(hostname string) error {

	// if CheckIfRegisterd() {
	// 	log.Printf("Node already registered")
	// 	return nil
	// }
	log.Printf("Registering node")
	nodeIP, err := GetPrimaryIPv4Address()
	if err != nil {
		return err
	}

	node := apiobject.Node{
		APIObject: apiobject.APIObject{
			APIVersion: configs.API_VERSION,
			Kind:       apiobject.NodeKind,
			Metadata: apiobject.Metadata{
				Name: hostname,
			},
		},
		Spec: apiobject.NodeSpec{IP: nodeIP},
	}

	url := configs.GetApiServerUrl() + configs.NodesUrl
	jsonData, _ := json.Marshal(node)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	return nil

}

// StartServer starts the Kubelet server to manage pods
func (k *Kubelet) WatchPods() {

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {

		err := k.MonitorAndManagePods()
		if err != nil {
			log.Printf("KUBELET error: %v", err)
			continue
		}
	}
}
