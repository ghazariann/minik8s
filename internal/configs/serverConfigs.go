package configs

const (
	LOCAL_API_SERVER_IP  = "127.0.0.1"
	MASTER_API_SERVER_IP = "192.168.1.13" // vahag_node
	API_SERVER_PORT      = "8081"
	SCHEMA               = "http://"
	LOCAL_API_URL        = SCHEMA + LOCAL_API_SERVER_IP + ":" + API_SERVER_PORT
	MASTER_API_URL       = SCHEMA + MASTER_API_SERVER_IP + ":" + API_SERVER_PORT
	masterNode           = true
)

func GetApiServerUrl() string {
	if masterNode {
		return LOCAL_API_URL
	} else {
		return MASTER_API_URL
	}
}

const (
	NodesURL          = "/nodes"
	NodeSpecURL       = "/api/v1/nodes/:name"
	NodeSpecStatusURL = "/api/v1/nodes/:name/status"
	NodeAllPodsURL    = "/api/v1/nodes/:name/pods"

	PodUrl      = "/pod"
	PodStoreUrl = "/podStore"
	PodsURL     = "/pods"

	ServiceURL      = "/service"
	ServicesURL     = "/services"
	ServiceStoreURL = "/serviceStore"
	DeploymentUrl   = "/deployment"
	DeploymentsUrl  = "/deployments"
	EndpointURL     = "/endpoints"
)

// create a path for the etcd key
// var SERVICES_URL = path.Join(API_URL, ETCDServicePath)
// var SERVICE_URL = path.Join(API_URL, ETCDServicePath)
// var DEPLOYMENTS_URL = path.Join(API_URL, ETCDDeploymentPath)
// var PODS_URL = path.Join(API_URL, ETCDPodPath)
// var ENDPOINTS_URL = path.Join(API_URL, ETCDServicePath)
