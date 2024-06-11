#!/bin/bash

# Create the deployment using kubectl and capture the output
output_create=$(./kubectl create deployment -f testdata/deployment.yaml)
expected_output_create="Deployment created: simple-deployment"

# Check if the deployment was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "Deployment creation output matched expected output."
else
    echo "Deployment creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Function to check deployment status
check_deployment_status() {
    for attempt in {1..5}; do
        echo "Checking deployment status, attempt $attempt of 5..."
        output_deployments=$(./kubectl get deployments)
        ready_replicas=$(echo "$output_deployments" | grep "simple-deployment" | awk '{print $3}')
        if [ "$ready_replicas" -gt 0 ]; then
            echo "Deployment has at least one ready replica."
            return 0
        fi
        sleep 5
    done
    echo "Deployment status check failed or no replicas are ready after multiple attempts."
    echo "Last output: $output_deployments"
    return 1
}

# Check deployment status with retries
if ! check_deployment_status; then
    exit 1
fi
./kubectl delete deployment simple-deployment
sleep 5
echo "All deployment checks passed!"
