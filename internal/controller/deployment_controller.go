package controller

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

func GetPodsFromAPIServer() ([]apiobject.PodStore, error) {
	url := configs.API_URL + "/pods"

	allPods := make([]apiobject.PodStore, 0)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list pods: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &allPods) != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	return allPods, nil
}

func GetAllDeploymentsFromAPIServer() ([]apiobject.DeploymentStore, error) {
	url := configs.API_URL + "/deployments"

	allDeployments := make([]apiobject.DeploymentStore, 0)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list deployments: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if json.Unmarshal(body, &allDeployments) != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	return allDeployments, nil
}

func FilterBySelector(pod *apiobject.PodStore, selectors map[string]string) bool {
	podLabel := pod.Metadata.Labels
	for key, value := range selectors {
		if podLabel[key] != value {
			return false
		} else {
			continue
		}
	}

	return true
}
func RandomStr(length int) string {
	var str string
	for i := 0; i < length; i++ {
		str += string(rune(rand.Intn(26) + 97))
	}
	return str
}
func AddReplica(deploymentMeta *apiobject.Metadata, pod *apiobject.PodTemplate, num int) error {
	url := configs.API_URL + "/pods"
	newPod := apiobject.Pod{}
	newPod.Metadata = pod.Metadata
	newPod.Kind = "Pod"
	newPod.APIVersion = "v1"
	newPod.Spec = pod.Spec
	// newPod.Metadata.Labels["deployement_name"] = deploymentMeta.Name
	// newPod.Metadata.Labels["deployement_namespace"] = deploymentMeta.Namespace
	// newPod.Metadata.Labels["deployement_uuid"] = deploymentMeta.UUID

	originalPodName := deploymentMeta.Name

	originalContainerNames := make([]string, 0)
	for _, container := range newPod.Spec.Containers {
		originalContainerNames = append(originalContainerNames, container.Name)
	}

	errStr := ""
	for i := 0; i < num; i++ {
		newPod.Metadata.Name = originalPodName + "-" + RandomStr(5)

		for index := range newPod.Spec.Containers {
			newPod.Spec.Containers[index].Name = originalContainerNames[index] + "-" + RandomStr(5)
		}
		jsonData, _ := json.Marshal(newPod)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			errStr += err.Error()
		}
	}

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}
func ReduceReplica(pods []apiobject.PodStore, num int) error {
	errStr := ""
	for i := 0; i < num; i++ {
		// choose a pod to delete randomly
		pod := pods[rand.Intn(len(pods))]
		url := fmt.Sprintf(configs.API_URL+"/pod?name=%s", pod.Metadata.Name)
		req, _ := http.NewRequest("DELETE", url, nil)
		_, err := http.DefaultClient.Do(req)
		if err != nil {
			errStr += err.Error()
		}
	}

	if errStr != "" {
		return errors.New(errStr)
	}

	return nil
}
func UpdateDeploymentStatus(filteredPods []apiobject.PodStore, deployment *apiobject.DeploymentStore) error {
	ReadyNums := 0
	for _, pod := range filteredPods {

		if pod.Status.Phase == apiobject.PodRunning {
			ReadyNums += 1
		}
	}

	if deployment.Status.ReadyReplicas == ReadyNums {
		return nil
	}
	deployment.Status.ReadyReplicas = ReadyNums
	url := fmt.Sprintf(configs.API_URL+"/deployment?name=%s", deployment.Metadata.Name)

	jsonData, _ := json.Marshal(deployment)
	_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	return nil
}
func DeletePodsByLabel() {

}
func DeploymenRoutine() {
	pods, err := GetPodsFromAPIServer()
	if err != nil {
		return
	}
	currentDeployments, err := GetAllDeploymentsFromAPIServer()

	if err != nil {
		return
	}
	// Detect deleted deployments and delete corresponding pods
	for _, prevDeployment := range previousDeployments {
		found := false
		for _, currDeployment := range currentDeployments {
			if prevDeployment.Metadata.UUID == currDeployment.Metadata.UUID {
				found = true
				break
			}
		}
		if !found {
			filteredPodsWithoutDeployment := make([]apiobject.PodStore, 0)
			for _, pod := range pods {
				if FilterBySelector(&pod, prevDeployment.Spec.Selector.MatchLabels) {
					filteredPodsWithoutDeployment = append(filteredPodsWithoutDeployment, pod)
				}
			}
			ReduceReplica(filteredPodsWithoutDeployment, len(filteredPodsWithoutDeployment))
		}
	}

	// Update previous deployments for the next iteration
	previousDeployments = currentDeployments

	for _, dp := range currentDeployments {
		filteredPods := make([]apiobject.PodStore, 0)
		for _, pod := range pods {
			if FilterBySelector(&pod, dp.Spec.Selector.MatchLabels) {
				filteredPods = append(filteredPods, pod)
			}
		}
		if len(filteredPods) < dp.Spec.Replicas {
			AddReplica(&dp.Metadata, &dp.Spec.Template, dp.Spec.Replicas-len(filteredPods))
		} else if len(filteredPods) > dp.Spec.Replicas {
			ReduceReplica(filteredPods, len(filteredPods)-dp.Spec.Replicas)
		}
		UpdateDeploymentStatus(filteredPods, &dp)
	}

}

var previousDeployments []apiobject.DeploymentStore

func WatchDeployment() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		DeploymenRoutine()
	}
}
