# Contributing

Install Golang v1.16 and have a Kubernetes cluster ready (e.g. Docker for Desktop).

Start:

```
make start
```

Start example using k3d Kubernetes cluster:

```bash
k3d cluster create argo-dataflow-test
kubectl config set-context k3d-argo-dataflow-test
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
GOBIN=$(pwd)/ GO111MODULE=on go get sigs.k8s.io/kustomize/kustomize/v3
mv kustomize /go/bin
```

Also required protobuf-compiler, python3 & pip3.

## Docker for Desktop and K3D Known Limitations

* Docker for Desktop
    * Does not enforce Kubernetes RBAC.
* K3D:
    * Requires you to import images.
    * Does not enforce resource requests.