package apiserver

import (
	"fmt"
	"log"
	"net/http"

	//"bytes"
	//"io/ioutil"
	"minik8s/internal/apiserver/handlers"
	"minik8s/internal/configs"
)

// TODO convert to gin
func StartServer() {
	// http.HandleFunc("/pod", handlers.HandlePods)
	http.HandleFunc(configs.PodUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetPod(w, r)
		case "DELETE":
			handlers.DeletePod(w, r)
		case "POST":
			handlers.UpdatePod(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	// http.HandleFunc("/pods", handlers.HandlePods)
	http.HandleFunc(configs.PodsURL, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetPods(w, r)
		case "POST":
			handlers.AddPod(w, r)

		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	// podStore
	http.HandleFunc(configs.PodStoreUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlers.UpdatePodStatus(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc(configs.ServiceURL, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetService(w, r)
		case "DELETE":
			handlers.DeleteService(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc(configs.ServicesURL, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetServices(w, r)
		case "POST":
			handlers.AddService(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc(configs.ServiceStoreURL, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlers.UpdateServiceStatus(w, r)
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	// deployment
	http.HandleFunc(configs.DeploymentUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetDeployment(w, r)
		case "DELETE":
			handlers.DeleteDeployment(w, r)
		case "POST":
			handlers.UpdateDeployment(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc(configs.DeploymentsUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetDeployments(w, r)
		case "POST":
			handlers.AddDeployment(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})

	// hpa
	http.HandleFunc(configs.HpaUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetHpa(w, r)
		case "DELETE":
			handlers.DeleteHpa(w, r)
		case "POST":
			handlers.UpdateHpa(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	// hpaStore
	http.HandleFunc(configs.HpaStoreUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlers.UpdateHpaStatus(w, r)
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc(configs.HpasUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetHpas(w, r)
		case "POST":
			handlers.AddHpa(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})

	// node
	http.HandleFunc(configs.NodeUrl, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetNode(w, r)
		case "DELETE":
			handlers.DeleteNode(w, r)
		case "POST":
			handlers.UpdateNode(w, r)
		default:

			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	// nodes
	http.HandleFunc(configs.NodesURL, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetNodes(w, r)
		case "POST":
			handlers.AddNode(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)

		}
	})

	fmt.Println("API Server starting on port " + configs.API_SERVER_PORT + " ...")
	if err := http.ListenAndServe(":"+configs.API_SERVER_PORT, nil); err != nil {
		log.Fatal(err)
	}
}
