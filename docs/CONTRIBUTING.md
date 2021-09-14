# Contributing

Install Golang v1.16 and have a Kubernetes cluster ready (e.g. Docker for Desktop).

Start:

```
make start
```

Start example using k3d Kubernetes cluster:

```bash
k3d cluster create
kubectl config set-context k3d-k3s-default
make start
```

To access the user interface, you must have checked out Argo Workflows in `../../argoproj/argo-workflows`. The UI will
appear on port 8080.

Before committing:

```
make pre-commit
```

Required dependencies:

```
GO111MODULE=off go get k8s.io/apimachinery           
GO111MODULE=off go get k8s.io/client-go
GO111MODULE=off go get k8s.io/api
GO111MODULE=off go get k8s.io/utils
GO111MODULE=off go get sigs.k8s.io/controller-runtime
GO111MODULE=off go get github.com/gogo/protobuf
```

Also required [protobuf-compiler](https://grpc.io/docs/protoc-installation/), python3 & pip3.

## Docker for Desktop and K3D Known Limitations

* Docker for Desktop
    * Does not enforce Kubernetes RBAC.
* K3D:
    * Requires you to import images.
    * Does not enforce resource requests.