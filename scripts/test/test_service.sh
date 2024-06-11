#!/bin/bash

# Create the service using kubectl and capture the output
output_create=$(./kubectl create service -f testdata/service.yaml)
expected_output_create="Service created: simple-service"

# Check if the service was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "Service creation output matched expected output."
else
    echo "Service creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Function to check service status
check_service_status() {
    for attempt in {1..5}; do
        echo "Checking service status, attempt $attempt of 5..."
        output_services=$(./kubectl get services)
        if [[ "$output_services" == *"simple-service"* && "$output_services" == *"running"* ]]; then
            echo "Service is running."
            # Extract the IP address of simple-service
            service_ip=$(echo "$output_services" | grep "simple-service" | awk '{print $3}')
            echo "IP Address of simple-service: $service_ip"
            return 0
        fi
        sleep 5
    done
    echo "Service status check failed or output did not match after multiple attempts."
    echo "Last output: $output_services"
    return 1
}

# Check service status with retries
if ! check_service_status; then
    exit 1
fi

# Assuming the function check_service_status found the service and extracted the IP
if [[ -z "$service_ip" ]]; then
    echo "Failed to extract IP address of simple-service."
    exit 1
fi

# Perform a curl request to the service
response=$(curl -s $service_ip)
expected_response_contains="<title>My Website</title>"

# Check if the response from the service is as expected
if [[ "$response" == *"$expected_response_contains"* ]]; then
    echo "Service response is as expected."
else
    echo "Service response did not match the expected content."
    echo "Expected to contain: $expected_response_contains"
    echo "Got: $response"
    exit 1
fi

# Clean up the service after the test
cleanup_output=$(./kubectl delete service simple-service)
expected_cleanup_output="Service deleted: simple-service"

if [[ "$cleanup_output" == "$expected_cleanup_output" ]]; then
    echo "Service cleanup was successful."
else
    echo "Failed to clean up the service."
    echo "Expected: $expected_cleanup_output"
    echo "Got: $cleanup_output"
    exit 1
fi
# ./kubectl delete service simple-service
# sleep 5
echo "All service checks passed!"
