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

# Wait for 10 seconds to let the pod start
sleep 15

# Check pod status using kubectl
output_pods=$(./kubectl get pods)
expected_output_pods="greet-pod                      running    10.32.0.1  26s                  vahag-master"

# Check if the pod status output is correct
if [[ "$output_pods" == *"$expected_output_pods"* ]]; then
    echo "Pod status output matched expected output."
else
    echo "Pod status check failed or output did not match."
    echo "Expected to contain: $expected_output_pods"
    echo "Got: $output_pods"
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

echo "Delete pod"
./kubectl delete pod greet-pod
sleep 5
echo "All checks passed!"


