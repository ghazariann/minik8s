package kubectl

import (
	"fmt"
	"io"
	"log"
	"minik8s/internal/configs"
	"net/http"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
}

var CmdDeletePod = &cobra.Command{
	Use:   "pod [name]",
	Short: "Delete the pod by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		DeletePod(args[0])
	},
}
var CmdDeleteDeployment = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Delete the deployment by name",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteDeployment(args[0])
	},
}

var CmdDeleteService = &cobra.Command{
	Use:  "service [name]",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		deleteService(args[0])

	},
}

func deleteService(name string) {
	url := fmt.Sprintf(configs.API_URL+"/service?name=%s", name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
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

// DeletePod deletes a pod by name
func DeletePod(name string) {
	url := fmt.Sprintf(configs.API_URL+"/pod?name=%s", name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
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

func deleteDeployment(name string) {
	url := fmt.Sprintf(configs.API_URL+"/deployment?name=%s", name)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
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

func init() {
	DeleteCmd.AddCommand(CmdDeletePod)
	DeleteCmd.AddCommand(CmdDeleteDeployment)
	DeleteCmd.AddCommand(CmdDeleteService)
}
