package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"minik8s/internal/apiobject"
	"net/http"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var CmdCreatePodFromYAML = &cobra.Command{
	Use:   "create pod -f [filename]",
	Short: "Create a pod from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		CreatePodFromYAML(filename)
	},
}

var CmdCreateServiceFromYAML = &cobra.Command{
	Use:   "create service -f [filename]",
	Short: "Create a service from a YAML file",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		CreateServiceFromYAML(filename)
	},
}

func CreatePodFromYAML(filename string) {
	data, err := ioutil.ReadFile(filename)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))

}

func CreateServiceFromYAML(filename string) {
	data, err := ioutil.ReadFile(filename)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}

func init() {
	CmdCreatePodFromYAML.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdCreatePodFromYAML.MarkFlagRequired("filename")

	CmdCreateServiceFromYAML.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdCreateServiceFromYAML.MarkFlagRequired("filename")
}
