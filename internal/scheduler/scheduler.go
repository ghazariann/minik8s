package scheduler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"
	"time"
)

var globCount int

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

func RoundRobin(nodes []apiobject.NodeStore) string {
	if len(nodes) == 0 {
		return ""
	}
	idx := globCount % len(nodes)
	globCount++
	return nodes[idx].Metadata.Name
}
func Random(nodes []apiobject.NodeStore) string {
	if len(nodes) == 0 {
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))
	idx := r.Intn(len(nodes))
	return nodes[idx].Metadata.Name
}
func ChooseNode(nodes []apiobject.NodeStore) string {
	if len(nodes) == 0 {
		return ""
	}
	switch configs.SchedulePolicy {
	case "RoundRobin":
		return RoundRobin(nodes)
	case "Random":
		return Random(nodes)
	default:
	}
	return ""
}
func GetAllNodes() ([]apiobject.NodeStore, error) {
	url := configs.GetApiServerUrl() + configs.NodesUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list nodes: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var nodes []apiobject.NodeStore
	if err := json.Unmarshal(body, &nodes); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	return nodes, nil
}
func SchedulePod(podStore *apiobject.PodStore) (string, error) {

	nodes, err := GetAllNodes()

	if err != nil {
		log.Fatal("NO NODES AVAILABLE")
		return "", err
	}

	var scheduledNode string

	if podStore.Spec.NodeName != "" {
		for _, node := range nodes {
			if node.Metadata.Name == podStore.Spec.NodeName {
				scheduledNode = podStore.Spec.NodeName
			}
		}
	}

	if scheduledNode == "" {
		scheduledNode = ChooseNode(nodes)
	}

	if scheduledNode == "" {
		return "", errors.New("no node available")
	}

	podStore.Spec.NodeName = scheduledNode
	// update
	// UpdatePodStatus(podStore)
	return scheduledNode, nil
}
