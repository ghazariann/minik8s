#!/bin/bash

# Create the pod using kubectl and capture the output
output_create=$(./kubectl create pod -f testdata/pod.yaml)
expected_output_create="Pod created: greet-pod"

# Check if the pod was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "Pod creation output matched expected output."
else
    echo "Pod creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Function to check pod status
check_pod_status() {
    for attempt in {1..5}; do
        echo "Checking pod status, attempt $attempt of 5..."
        output_pods=$(./kubectl get pods)
        if [[ "$output_pods" == *"greet-pod"* && "$output_pods" == *"running"* ]]; then
            echo "Pod is running."
            return 0
        fi
        sleep 5
    done
    echo "Pod status check failed or output did not match after multiple attempts."
    echo "Last output: $output_pods"
    return 1
}

# Check pod status with retries
if ! check_pod_status; then
    exit 1
fi

# Check the running containers with docker ps
output_docker_ps=$(docker ps --format "{{.Names}}")
expected_containers=("greet-pod_dir-creator" "greet-pod_welcome-container" "greet-pod_greet-container" "greet-pod_pause")

# Validate all expected containers are running
for container in "${expected_containers[@]}"; do
    if [[ "$output_docker_ps" == *"$container"* ]]; then
        echo "$container is running."
    else
        echo "Container $container is not running."
        exit 1
    fi
done
# ./kubectl delete pod greet-pod
# sleep 5

echo "All checks passed!"
