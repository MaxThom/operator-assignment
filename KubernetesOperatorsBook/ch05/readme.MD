kubectl apply -f database.yaml
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml

kubectl port-forward service/visitors-frontend-service 30686:3000
kubectl port-forward service/visitors-backend-service 30685:8000

Navigate to: http://127.0.0.1:30686/

kubectl delete -f database.yaml
kubectl delete -f backend.yaml
kubectl delete -f frontend.yaml

