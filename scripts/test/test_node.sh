#!/bin/bash

# Create the node using kubectl and capture the output
output_create=$(./kubectl create pod -f testdata/node.yaml)
expected_output_create="Pod created: vahag-node"

# Check if the node was created successfully
if [[ "$output_create" == "$expected_output_create" ]]; then
    echo "Node creation output matched expected output."
else
    echo "Node creation failed or output did not match."
    echo "Expected: $expected_output_create"
    echo "Got: $output_create"
    exit 1
fi

# Check node status using kubectl
output_nodes=$(./kubectl get nodes)
expected_output_nodes="192.168.1.18     vahag-master running    1          0.791623   0.006705"

# Check if the node status output is correct
if [[ "$output_nodes" == *"$expected_output_nodes"* ]]; then
    echo "Node status output matched expected output."
else
    echo "Node status check failed or output did not match."
    echo "Expected to contain: $expected_output_nodes"
    echo "Got: $output_nodes"
    exit 1
fi
./kubectl delete node vahag-node
sleep 5
echo "All node checks passed!"
