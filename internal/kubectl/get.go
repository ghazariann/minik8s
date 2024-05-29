package kubectl

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var GetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources",
}

var CmdGetPod = &cobra.Command{
	Use:   "pod [name]",
	Short: "Retrieve information about the pod by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetPod(args[0])
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
	Run: func(cmd *cobra.Command, args []string) {
		ListDeployments()
	},
}

func GetPod(name string) {
	url := fmt.Sprintf("http://localhost:8080/pods?name=%s", name)
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

func GetAllPods() {
	url := "http://localhost:8080/all-pods"
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
	url := fmt.Sprintf("http://localhost:8080/services?name=%s", name)
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
	url := "http://localhost:8080/all-services"
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

func ListDeployments() {
	url := "http://localhost:8080/deployments"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error sending request to list deployments: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func init() {
	GetCmd.AddCommand(CmdGetDeployment)
	GetCmd.AddCommand(CmdGetDeployments)
	GetCmd.AddCommand(CmdGetAllServices)
	GetCmd.AddCommand(CmdGetService)
	GetCmd.AddCommand(CmdGetPod)
	GetCmd.AddCommand(CmdGetAllPods)
}
