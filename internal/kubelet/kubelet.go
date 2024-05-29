package kubelet

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

type Kubelet struct {
	Name           string
	RuntimeManager *RuntimeManager
}

// NewKubelet initializes and returns a new Kubelet
func NewKubelet() (*Kubelet, error) {
	dockerClient, err := NewDockerClient()
	if err != nil {
		return nil, err
	}
	runtimeManager := NewRuntimeManager(dockerClient)
	return &Kubelet{
		RuntimeManager: runtimeManager,
	}, nil
}
func (k *Kubelet) GetAllPods() ([]apiobject.PodStore, error) {
	url := configs.API_URL + "/pods"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
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
	for containerName, podUID := range knownContainers {
		if _, ok := newContainers[containerName]; !ok { // one or two does not matter
			// Container is missing, look for a matching pod
			if pod, podExists := pods[podUID]; podExists {
				// Pod exists, recreate the missing container
				log.Printf("Recreating missing container with name: %s", containerName)
				// list and find the container in pod.spec.containers
				k.RuntimeManager.CreatePod(&pod)
			}
		}
	}
}

func (k *Kubelet) MonitorAndManagePods() {
	var knownPods = map[string]apiobject.PodStore{} // UUID -> PodStore
	var knownContainers = map[string]string{}       // ContainerID -> PodUUID

	for {
		// Fetch all pods

		containers, _ := k.RuntimeManager.DockerClient.ListPodContainers()
		k.SyncContainers(knownContainers, containers, knownPods)
		knownContainers = containers

		pods, err := k.GetAllPods()

		if err != nil {
			log.Printf("Error fetching pods: %v", err)
			time.Sleep(10 * time.Second) // Wait before retrying
			continue
		}
		// Convert fetched pods to a map for easy comparison
		currentPods := map[string]apiobject.PodStore{}
		for _, pod := range pods {
			currentPods[pod.Metadata.UUID] = pod
		}

		// Detect deleted pods
		for podName := range knownPods {
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
		knownPods = currentPods

		// Process pending pods
		for _, pod := range pods {
			if pod.Status.Phase == apiobject.PodPending {
				log.Printf("Pod %s is pending. Attempting to create containers...", pod.Metadata.Name)
				if err := k.RuntimeManager.CreatePod(&pod); err != nil {
					log.Printf("Error creating pod %s: %v", pod.Metadata.Name, err)
				} else {
					log.Printf("Successfully created containers for pod %s", pod.Metadata.Name)
					// Update pod status to Running
					pod.Status.Phase = apiobject.PodRunning
					if err := k.UpdatePodStatus(&pod); err != nil {
						log.Printf("Error updating pod status for %s: %v", pod.Metadata.Name, err)
					} else {
						log.Printf("Successfully updated pod status for %s to Running", pod.Metadata.Name)
					}
				}
			}
		}

		time.Sleep(5 * time.Second) // Poll interval
	}
}

// CleanUpPod stops and removes all containers associated with the pod
func (k *Kubelet) CleanUpPod(podName string) error {
	log.Printf("Cleaning up resources for pod %s", podName)
	return k.RuntimeManager.DeletePod(podName)
}

// UpdatePodStatus sends a request to the API server to update the pod status
func (k *Kubelet) UpdatePodStatus(pod *apiobject.PodStore) error {
	url := fmt.Sprintf(configs.API_URL+"/podStore?name=%s", pod.Metadata.Name)
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

// StartServer starts the Kubelet server to manage pods
func (k *Kubelet) StartServer() {
	k.MonitorAndManagePods()
}
