#!/bin/bash

# Create the HPA using kubectl and capture the output
output_create=$(./kubectl create hpa -f testdata/hpa.yaml)
expected_output_create="Hpa created: test-hpa"

# Check if the HPA was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "HPA creation output matched expected output."
else
    echo "HPA creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Function to check HPA status
check_hpa_status() {
    for attempt in {1..5}; do
        echo "Checking HPA status, attempt $attempt of 5..."
        output_hpas=$(./kubectl get hpas)
        current_replicas=$(echo "$output_hpas" | grep "test-hpa" | awk '{print $4}')
        if [ "$current_replicas" -gt 0 ]; then
            echo "HPA is managing at least one pod."
            return 0
        fi
        sleep 5
    done
    echo "HPA status check failed or HPA is not managing any pods after multiple attempts."
    echo "Last output: $output_hpas"
    return 1
}

# Check HPA status with retries
if ! check_hpa_status; then
    exit 1
fi
./kubectl delete hpa test-hpa

echo "All HPA checks passed!"
