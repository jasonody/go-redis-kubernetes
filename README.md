# go-redis-kubernetes
Go app with Redis on Kubernetes

[Source](https://www.callicoder.com/deploy-multi-container-go-redis-app-kubernetes/)

## Deploy from image pulled from Docker Hub registry

#### Deploy
1. Redis: `kubectl apply -f deployment/redis-master.yaml`
2. API server: `kubectl apply -f deployment/go-redis-kubernetes.yaml`

#### Delete deployment
1. API server: `kubectl delete -f deployment/go-redis-kubernetes.yaml`
2. Redis: `kubectl delete -f deployment/redis-master.yaml`

## Testing Go app server
- Get the service's port as an env var: `kubectl get services | grep go-redis-kubernetes-service | egrep -o ':(.*?)\/' | grep '[0-9]'`
- Make an HTTP request: `curl localhost:<service's port>`