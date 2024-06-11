# Execute clean_apiobjects.sh
echo "Testing Pods"
./scripts/test/test_pod.sh
echo "Testing Deployments"
./scripts/test/test_deployment.sh
echo "Testing HPAs"
./scripts/test/test_hpa.sh
echo "Testing Services"
./scripts/test/test_service.sh
echo "Testing Nodes"
./scripts/test/test_nodes.sh
echo "Testing DNS"
./scripts/test/test_dns.sh