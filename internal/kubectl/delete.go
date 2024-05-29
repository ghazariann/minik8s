package kubectl

import (
	"fmt"
	"io"
	"log"
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

// DeletePod deletes a pod by name
func DeletePod(name string) {
	url := fmt.Sprintf("http://localhost:8080/pod?name=%s", name)
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
	url := fmt.Sprintf("http://localhost:8080/deployment?name=%s", name)
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
}
