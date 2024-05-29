# Mykubectl (Client) Commands

## Pods
- **Get Pod**: `getpod [name]` (GET: `http://localhost:8080/pods?name=%s`)
- **List All Pods**: `getallpod` (GET: `http://localhost:8080/all-pods`)
- **Create Pod**: `createpod -f [filename]` (POST: `http://localhost:8080/pods`)

## Services
- **Get Service**: `getservice [name]` (GET: `http://localhost:8080/services?name=%s`)
- **List All Services**: `getallservices` (GET: `http://localhost:8080/all-services`)
- **Create Service**: `createservice -f [filename]` (POST: `http://localhost:8080/services`)

## Deployments
- **Get Deployment**: `getdeployment [name]` (GET: `http://localhost:8080/deployments?name=%s`)
- **List All Deployments**: `getalldeployments` (GET: `http://localhost:8080/deployments`)
- **Create Deployment**: `createdeployment -f [filename]` (POST: `http://localhost:8080/deployments`)
- **Delete Deployment**: `deletedeployment [name]` (DELETE: `http://localhost:8080/deployments?name=%s`)
- **Update Deployment**: `updatedeployment -f [filename]` (PUT: `http://localhost:8080/deployments`)

# Apiserver Endpoint Handlers

## Pods
- **Handle Pods**: (GET, POST: `http://localhost:8080/pods`)
- **Handle All Pods**: (GET: `http://localhost:8080/all-pods`)
- **Handle Unscheduled Pods**: (GET: `http://localhost:8080/unscheduled-pods`)
- **Update Pod**: (POST: `http://localhost:8080/updatePod`)

## Services
- **Handle Services**: (GET, POST: `http://localhost:8080/services`)
- **Handle All Services**: (GET: `http://localhost:8080/all-services`)

## Scheduler Operations
- **Start Scheduler**: Periodically checks and schedules unscheduled pods every 10 seconds.

# Kubelet
- **Handle Start Pod**: (POST: `http://localhost:10250/startPod`)

# etcd Operations
- **Update Pod**: `pods/<pod-name>` (GET, PUT)
- **Get Pods by Condition**: `pods/` (GET)
- **Get, Put, and GetAll Keys Operations**
