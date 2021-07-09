## Rocket Lab Assignment
---
[MaxThom GitHub Profile](https://github.com/MaxThom)

### Important files
- TmSource Container: ```/tm/Dockerfile```
- k8s files: ```/k8s/config/```
- Operator: ```/k8s/controller/```
- Operators controllers: ```/k8s/controller/controllers/```
- Operators crds: ```/k8s/controller/api/v1/ (generated to yaml /k8s/controller/config/crd/base/)```
- Operators crs: ```/k8s/controller/config/samples/```

### Stack
- windows with WSL2
- k3d
- kubebuilder
- go

### Docker
- Used alpine for smallest image footprint.
- Depending on the use case and requirements, debian/ubuntu could be used.

### Assumptions
- When a site is disabled, their linked tmsource pods are deleted, but not the tmsource object.
- If a new tmsource is created and his site is disabled, his pod is not created.
- When a tmsource is deleted, his pod is as well.
- When a site is set to enable, all his linked tmsource pods are created.
- If a site is deleted, all his linked tmsources object and pods are deleted.
- If a tmsource config is changed, delete and recreate his pod.
- You can create tmsource even if their site does not exist.
- You can use metadata.name instead of spec.name to link site.

### Improvements
- Add index on spec.site to speed up query.
- Or set site in metadata.label to using built in query on labels and allow query from kubectl (ex: get all tmsources from a site).
- Use events and watchers to monitor pods.

### Some commands
#### Kubebuild
- go mod init github.com/maxthom/rocketlab-controller
- kubebuilder init --domain rocketlab.global
- kubebuilder create api --group tm --version v1 --kind Site
- kubebuilder create api --group tm --version v1 --kind TmSource
- make manifests
- make install
- make run
- make docker-build docker-push IMG=maxthom/rocket-controller:latest
- make deploy IMG=maxthom/rocket-controller:latest

#### K3d
- k3d cluster create dev-rocket --api-port 127.0.0.1:6445 -p 8080:80@loadbalancer
- kubectl port-forward --namespace default nats-server-deployment-64686d457b-z9qqf 4222:4222

#### Docker TmSource
- docker build -t maxthom/rocket-source .
- docker push maxthom/rocket-source:latest
- docker run -it --rm --name rocket-source -e METRIC_NAME=Rock -e NATS_SERVICE_PORT=<> maxthom/rocket-source
