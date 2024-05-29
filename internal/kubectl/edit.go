package kubectl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"minik8s/internal/apiobject"
	"minik8s/internal/configs"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var EditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit resources",
}

var CmdEditPod = &cobra.Command{
	Use:   "pod [name] -f [filename]",
	Short: "Edit the pod by name",
	Run: func(cmd *cobra.Command, args []string) {
		filename, _ := cmd.Flags().GetString("filename")
		EditPod(args[0], filename)
	},
}
var CmdEditDeployment = &cobra.Command{
	Use:   "deployment [name] -f [filename]",
	Short: "Edit the deployment by name",
	Run: func(cmd *cobra.Command, args []string) {
		deploymentname, _ := cmd.Flags().GetString("name")
		filename, _ := cmd.Flags().GetString("filename")
		EditDeployment(deploymentname, filename)
	},
}

func EditDeployment(name string, filename string) {
}

func EditPod(name string, filename string) {
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

	url := fmt.Sprintf(configs.GetApiServerUrl()+configs.PodUrl+"?name=%s", name)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch pod: %s\n", resp.Status)
		return
	}
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}
	fmt.Println("Pod updated")
}

func init() {
	CmdEditPod.Flags().StringVarP(new(string), "filename", "f", "", "Path to the YAML file")
	CmdEditPod.MarkFlagRequired("filename")
	CmdCreateDeployment.MarkFlagRequired("filename")
	EditCmd.AddCommand(CmdEditPod)
	EditCmd.AddCommand(CmdEditDeployment)
}
