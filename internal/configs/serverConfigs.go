package configs

const (
	LOCAL_API_SERVER_IP = "127.0.0.1"
	API_SERVER_PORT     = "8081"
	API_URL             = "http://" + LOCAL_API_SERVER_IP + ":" + API_SERVER_PORT
)

// create a path for the etcd key
// var SERVICES_URL = path.Join(API_URL, ETCDServicePath)
// var SERVICE_URL = path.Join(API_URL, ETCDServicePath)
// var DEPLOYMENTS_URL = path.Join(API_URL, ETCDDeploymentPath)
// var PODS_URL = path.Join(API_URL, ETCDPodPath)
// var ENDPOINTS_URL = path.Join(API_URL, ETCDServicePath)
