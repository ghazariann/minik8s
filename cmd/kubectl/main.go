package main

import (
	"log"

	"minik8s/internal/kubectl"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kubectl",
		Short: "kubectl interacts with a Kubernetes-like API server",
	}

	rootCmd.AddCommand(
		kubectl.CmdGetPod,
		kubectl.CmdGetAllPods,
		kubectl.CmdCreatePodFromYAML,

		kubectl.CmdGetService,
		kubectl.CmdGetAllServices,
		kubectl.CmdCreateServiceFromYAML,
	)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing kubectl: %v", err)
	}
}
