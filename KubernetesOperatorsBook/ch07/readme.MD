OPERATOR_NAME=visitors-operator
operator-sdk init $OPERATOR_NAME --plugins go/v2 --domain example.org --owner "MaxThom" --repo=github.com/maxthom

operator-sdk create api --group cache --version v1alpha1 --kind VisitorsApp --resource --controller

# Modify visitorsapp_types.go
make generate
make manifests

# Run
In debug,
make install run
kubectl apply -f config/samples/cache_v1alpha1_visitorsapp.yaml

Inside cluster,
make deploy
kubectl apply -f config/samples/cache_v1alpha1.
_visitorsapp.yaml

kubectl get visitorsapps
kubectl describe visitorsapps visitorsapp-sample