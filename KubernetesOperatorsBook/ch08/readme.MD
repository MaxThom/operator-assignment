kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.2/crds.yaml

kubectl apply -f https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.18.2/olm.yaml

kubectl get pods -n olm
kubectl get packagemanifest -n olm

apply all-og.yaml & sub.yaml

kubectl get csv -n default
kubectl get sub -n default
kubectl describe csv/etcdoperator.v0.9.4 -n default

kubectl delete csv/etcdoperator.v0.9.4
kubectl delete subscription/etcd-subscription

# To generate bundle
make bundle

https://sdk.operatorframework.io/docs/olm-integration/generation/#generate-your-first-release

git clone https://github.com/operator-framework/operator-marketplace.git
cd operator-marketplace
kubectl apply -f deploy/upstream/
kubectl get ns marketplace

pip3 install operator-courier