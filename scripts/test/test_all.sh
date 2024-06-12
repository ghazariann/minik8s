# Execute clean_apiobjects.sh
echo "Testing Pods"
./scripts/test/test_pod.sh
echo "Testing Deployments"
./scripts/test/test_deployment.sh

echo "Testing Services"
./scripts/test/test_service.sh
echo "Testing Nodes"
./scripts/test/test_node.sh
echo "Testing HPAs"
./kubectl create pod -f testdata/podhpa.yaml
sleep 10
./scripts/test/test_hpa.sh
echo "Testing DNS"
./scripts/test/test_dns.sh