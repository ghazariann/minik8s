package apiserver

import (
	"fmt"
	"log"
	"net/http"

	//"bytes"
	//"io/ioutil"

	"minik8s/internal/apiserver/handlers"
)

// TODO convert to gin
func StartServer() {
	// http.HandleFunc("/pods", handlers.HandlePods)
	http.HandleFunc("/pods", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetPods(w, r)
		case "POST":
			handlers.AddPod(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/podStore", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlers.UpdatePodStatus(w, r)
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/pod", func(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/services", handlers.HandleServices)
	http.HandleFunc("/all-services", handlers.HandleAllServices)
	http.HandleFunc("/deployments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetDeployments(w, r)
		case "POST":
			handlers.AddDeployment(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/deployment", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handlers.GetDeployment(w, r)
		case "DELETE":
			handlers.DeleteDeployment(w, r)
		case "PUT":
			handlers.UpdateDeployment(w, r)
		default:
			http.Error(w, "Unsupported HTTP method", http.StatusMethodNotAllowed)
		}
	})
	fmt.Println("API Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
