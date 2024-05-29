package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"minik8s/internal/apiobject"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
}

var CmdCreatePodFromYAML = &cobra.Command{
	Use:   "pod -f [filename]",
	Short: "Create a pod from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		CreatePodFromYAML(filename)
	},
}

var CmdCreateServiceFromYAML = &cobra.Command{
	Use:   "service -f [filename]",
	Short: "Create a service from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		CreateServiceFromYAML(filename)
	},
}

var CmdCreateDeployment = &cobra.Command{
	Use:   "deployment -f [filename]",
	Short: "Create a deployment from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		CreateDeploymentFromYAML(filename)
	},
}

func CreatePodFromYAML(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var pod apiobject.Pod
	if err := yaml.Unmarshal(data, &pod); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	jsonData, err := json.Marshal(pod)
	if err != nil {
		log.Fatalf("Error converting pod data to JSON: %v", err)
	}

	url := "http://localhost:8080/pods"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))

}

func CreateServiceFromYAML(filename string) {
	data, err := os.ReadFile(filename)
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

	url := "http://localhost:8080/services"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func CreateDeploymentFromYAML(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v", err)
	}

	var deployment apiobject.Deployment
	if err := yaml.Unmarshal(data, &deployment); err != nil {
		log.Fatalf("Error parsing YAML: %v", err)
	}

	jsonData, err := json.Marshal(deployment)
	if err != nil {
		log.Fatalf("Error converting deployment data to JSON: %v", err)
	}

	url := "http://localhost:8080/deployments"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request to create deployment: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func init() {
	CmdCreateDeployment.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdCreateDeployment.MarkFlagRequired("filename")

	CmdCreatePodFromYAML.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdCreatePodFromYAML.MarkFlagRequired("filename")

	CmdCreateServiceFromYAML.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdCreateServiceFromYAML.MarkFlagRequired("filename")

	CreateCmd.AddCommand(CmdCreatePodFromYAML)
	CreateCmd.AddCommand(CmdCreateDeployment)
	CreateCmd.AddCommand(CmdCreateServiceFromYAML)

}
