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
	http.HandleFunc("/pods", handlers.HandlePods)
	http.HandleFunc("/all-pods", handlers.HandleAllPods)
	http.HandleFunc("/unscheduled-pods", handlers.HandleUnscheduledPods)
	http.HandleFunc("/updatePod", handlers.HandleUpdatePod)
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
	fmt.Println("API Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
