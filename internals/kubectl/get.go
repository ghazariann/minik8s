package kubectl

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var CmdGetPod = &cobra.Command{
	Use:   "get pod [name]",
	Short: "Retrieve information about the pod by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetPod(args[0])
	},
}

var CmdGetAllPods = &cobra.Command{
	Use:   "get pods",
	Short: "Retrieve information about all pods",
	Run: func(cmd *cobra.Command, args []string) {
		GetAllPods()
	},
}

var CmdGetService = &cobra.Command{
	Use:   "get service [name]",
	Short: "Retrieve information about the service by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		GetService(args[0])
	},
}

var CmdGetAllServices = &cobra.Command{
	Use:   "get services",
	Short: "Retrieve information about all services",
	Run: func(cmd *cobra.Command, args []string) {
		GetAllServices()
	},
}

func GetPod(name string) {
	url := fmt.Sprintf("http://localhost:8080/pods?name=%s", name)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
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
	body, err := ioutil.ReadAll(resp.Body)
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
	body, err := ioutil.ReadAll(resp.Body)
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
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println(string(body))
}
