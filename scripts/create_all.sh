
./kubectl create deployment -f testdata/deployment.yaml
./kubectl create service -f testdata/service.yaml

echo "Waiting..."
sleep 10 
echo docker ps
docker ps
./kubectl get pods
./kubectl get services
# ./kubectl get endpoints
./kubectl get deployments
