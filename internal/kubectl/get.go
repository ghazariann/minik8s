package kubectl

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"

	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
}

var CmdGetAll = &cobra.Command{
	Use:   "all",
	Short: "Get all resources",
	Run: func(cmd *cobra.Command, args []string) {
		GetAllPods()
		GetAllServices()
		ListDeployments()
		ListHpas()
		ListNodes()
	},
}

var CmdGetPod = &cobra.Command{
	Use:   "pod [name]",
	Short: "Retrieve information about the pod by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetPod(args[0])
	},
}

var CmdGetEndpoints = &cobra.Command{
	Use:   "endpoints",
	Short: "Retrieve information about the endpoints",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetEndpoints()
	},
}
var CmdGetAllPods = &cobra.Command{
	Use:   "pods",
	Short: "Retrieve information about all pods",
	Run: func(cmd *cobra.Command, args []string) {
		GetAllPods()
	},
}

var CmdGetService = &cobra.Command{
	Use:   "service [name]",
	Short: "Retrieve information about the service by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetService(args[0])
	},
}

var CmdGetAllServices = &cobra.Command{
	Use:   "services",
	Short: "Retrieve information about all services",
	Run: func(cmd *cobra.Command, args []string) {
		GetAllServices()
	},
}

// CmdGetDeployments - Cobra command to list all deployments
var CmdGetDeployments = &cobra.Command{
	Use:   "deployments",
	Short: "List all deployments",
	Run: func(cmd *cobra.Command, args []string) {
		ListDeployments()
	},
}

var CmdGetDeployment = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Get One deployment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetDeployment(args[0])
	},
}

// implementation of the hpa get
var CmdGetHpa = &cobra.Command{
	Use:   "hpa [name]",
	Short: "Get One hpa",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetHpa(args[0])
	},
}

// implementation of the hpa get all
var CmdGetHpas = &cobra.Command{
	Use:   "hpas",
	Short: "List all hpas",
	Run: func(cmd *cobra.Command, args []string) {
		ListHpas()
	},
}

// node
var CmdGetNode = &cobra.Command{
	Use:   "node [name]",
	Short: "Get One node",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetNode(args[0])
	},
}

var CmdGetNodes = &cobra.Command{
	Use:   "nodes",
	Short: "List all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		ListNodes()
	},
}

func GetPod(name string) {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.PodUrl+"?name=%s", name)
	resp, err := http.Get(url)
	// Check for HTTP status code
	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Pod %s not found\n", name)
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch pod: %s\n", resp.Status)
		return
	}
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// Unmarshal the JSON response into a PodStore
	var podStore apiobject.PodStore
	if err := json.Unmarshal(body, &podStore); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	// Marshal with indentation for pretty printing
	formattedJSON, err := json.MarshalIndent(podStore, "", "    ")
	if err != nil {
		log.Fatalf("Error formatting JSON: %v", err)
	}

	fmt.Println(string(formattedJSON))
}

func GetDeployment(name string) {
	url := configs.GetApiServerUrl() + configs.DeploymentUrl + "?name=" + name
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Deployment %s not found\n", name)
		return
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	// Unmarshal the JSON response into a PodStore
	var dep apiobject.DeploymentStore
	if err := json.Unmarshal(body, &dep); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	// Marshal with indentation for pretty printing
	formattedJSON, err := json.MarshalIndent(dep, "", "    ")
	if err != nil {
		log.Fatalf("Error formatting JSON: %v", err)
	}

	fmt.Println(string(formattedJSON))
}
func GetAllPods() {
	url := configs.GetApiServerUrl() + configs.PodsURL
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
	var pods []apiobject.PodStore
	if err := json.Unmarshal(body, &pods); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}
	// Print header
	fmt.Printf("%-30s %-10s %-10s %-20s %-10s\n", "Name", "Status", "IP", "Node", "Runtime")

	// Print each container's name and status
	for _, pod := range pods {
		fmt.Printf("%-30s %-10s %-10s %-20s %-10s\n", pod.Metadata.Name, pod.Status.Phase, pod.Status.PodIP, pod.Spec.NodeName, pod.Status.UpdateTime)
	}
	// // Marshal with indentation for pretty printing
	// formattedJSON, err := json.MarshalIndent(pods, "", "    ")
	// if err != nil {
	// 	log.Fatalf("Error formatting JSON: %v", err)
	// }

	// fmt.Println(string(formattedJSON))
}

func GetEndpoints() {
	url := configs.GetApiServerUrl() + configs.EndpointsURL
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))

}
func GetService(name string) {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.ServiceUrl+"?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func GetAllServices() {
	url := configs.GetApiServerUrl() + configs.ServicesUrl
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

	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("%-15s  %-10s   %-15s %-30s\n", "Name", "Phase", "Cluster IP", "Endpoints")

	// Print each container's name and status
	for _, service := range services {
		// get endpoints in string with ; concat
		endpoints := ""
		for _, endpoint := range service.Status.Endpoints {
			endpoints = endpoints + " " + endpoint.IP + "/" + endpoint.Ports[0]
		}
		fmt.Printf("%-15s %-10s %-15s %-30s \n", service.Metadata.Name, service.Status.Phase, service.Spec.ClusterIP, endpoints)
	}
}

func ListDeployments() {
	url := configs.GetApiServerUrl() + configs.DeploymentsUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list deployments: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var deployments []apiobject.DeploymentStore
	if err := json.Unmarshal(body, &deployments); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("%-10s  %-10s  %-10s\n", "Name", "Replicas", "Ready Replicas")

	// Print each container's name and status
	for _, deployment := range deployments {
		fmt.Printf("%-10s %-10d  %-10d\n", deployment.Metadata.Name, deployment.Spec.Replicas, deployment.Status.ReadyReplicas)
	}
}

func GetHpa(name string) {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.HpaUrl+"?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func ListHpas() {
	url := configs.GetApiServerUrl() + configs.HpasUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list hpas: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	var hpas []apiobject.HpaStore
	if err := json.Unmarshal(body, &hpas); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	fmt.Printf("%-10s  %-10s %-10s %-10s %-10s %-10s\n", "Name", "Min Replicas", "Max Replicas", "Current Replicas", "currentCPUPercent", "currentMemPercent")

	// Print each container's name and status
	for _, hpa := range hpas {
		fmt.Printf("%-10s %-10d %-10d %-10d %-10f %-10f\n", hpa.Metadata.Name, hpa.Spec.MinReplicas, hpa.Spec.MaxReplicas, hpa.Status.CurrentReplicas, hpa.Status.CurrentCPUPercent, hpa.Status.CurrentMemPercent)
	}
}

func GetNode(name string) {
	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.NodeUrl+"?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func ListNodes() {
	url := configs.GetApiServerUrl() + configs.NodesUrl
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list nodes: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	var nodes []apiobject.NodeStore
	if err := json.Unmarshal(body, &nodes); err != nil {
		log.Fatalf("Error unmarshalling response body: %v", err)
	}

	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// print Spec.IP, Status.HostName, Status.Condition, Status.NumPods, Status.CpuPercent, Status.MemPercent
	fmt.Printf("%-15s  %-10s  %-10s %-10s %-10s %-10s\n", "IP", "HostName", "Condition", "NumPods", "CpuPercent", "MemPercent")
	// Print each container's name and status
	for _, node := range nodes {
		fmt.Printf(" %-15s  %-10s %-10s %-10d %-10f %-10f\n", node.Spec.IP, node.Metadata.Name, node.Status.Condition, node.Status.NumPods, node.Status.CpuPercent, node.Status.MemPercent)
	}

}
func init() {
	GetCmd.AddCommand(CmdGetDeployment)
	GetCmd.AddCommand(CmdGetDeployments)
	GetCmd.AddCommand(CmdGetAllServices)
	GetCmd.AddCommand(CmdGetService)
	GetCmd.AddCommand(CmdGetPod)
	GetCmd.AddCommand(CmdGetAllPods)
	GetCmd.AddCommand(CmdGetHpa)
	GetCmd.AddCommand(CmdGetHpas)
	GetCmd.AddCommand(CmdGetNode)
	GetCmd.AddCommand(CmdGetNodes)
	GetCmd.AddCommand(CmdGetAll)
	GetCmd.AddCommand(CmdGetEndpoints)
}
