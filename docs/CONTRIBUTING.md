# Contributing

Install :

* Golang v1.16
* Docker
* Kubernetes cluster ready (e.g. Docker for Desktop).
* Python 3 and Pip 3 (optional)

Start:

```
make start
```

To start the user interface, you must have checked out Argo Workflows in `../../argoproj/argo-workflows`. The UI will
appear on port 8080, then run `make start UI=true`.

Before committing, of if you need to re-create codegen, then you need to run the pre-commit checks.

**Pre-commit in Docker (recommended)** 

One time set-up:

```bash
make make-image
```

Then:

```bash
./hack/make-in-docker.sh pre-commit -B
```

**Pre-commit locally**

One time set-up:

Install [protobuf-compiler](https://grpc.io/docs/protoc-installation/).

Then:

```bash
make pre-commit -B
```

## Docker for Desktop and K3D Known Limitations

* Docker for Desktop
    * Does not enforce Kubernetes RBAC.
* K3D:
    * Requires you to import images.
    * Does not enforce resource requests.
