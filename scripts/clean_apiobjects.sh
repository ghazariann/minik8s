#!/bin/bash

# Function to delete resources and print status
delete_resource() {
    resource_type=$1
    resource_names=$2

    echo "Deleting $resource_type ..."
    for resource_name in $resource_names; do
        ./kubectl delete $resource_type $resource_name
    done
}

# Clean up pods
pods=$(./kubectl get pods | awk 'NR>1 {print $1}')
echo $pods
delete_resource "pod" "$pods"

# Clean up services
services=$(./kubectl get services | awk 'NR>1 {print $1}')
delete_resource "service" "$services"

# Clean up deployments
deployments=$(./kubectl get deployments | awk 'NR>1 {print $1}')
delete_resource "deployment" "$deployments"

# Clean up HPAs
hpas=$(./kubectl get hpas | awk 'NR>1 {print $1}')
delete_resource "hpa" "$hpas"

# Check if resources are cleaned up
echo "Checking if resources are cleaned up ..."
sleep 3
echo "Pods:"
./kubectl get pods
echo "Services:"
./kubectl get services
echo "Deployments:"
./kubectl get deployments
echo "HPAs:"
./kubectl get hpa

