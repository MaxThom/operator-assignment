OPERATOR_NAME=visitors-helm-operator
operator-sdk init $OPERATOR_NAME --domain=example.com --version=v1 --kind=VisitorsApp --plugins=helm
wget https://github.com/kubernetes-operators-book/chapters/releases/download/1.0.0/visitors-helm.tgz
operator-sdk init $OPERATOR_NAME --domain=example.com --version=v1 --kind=VisitorsApp --plugins=helm --helm-chart=./visitors-helm.tgz
operator-sdk init $OPERATOR_NAME --domain=example.com --version=v1 --kind=VisitorsApp --plugins=ansible
