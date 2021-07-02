# Contributing

Install Golang v1.16 and have a Kubernetes cluster ready (e.g. Docker for Desktop).

Start:

```
make start
```

To access the user interface, you must have checked out Argo Workflows in `../../argoproj/argo-workflows`. The UI will
appear on port 8080.

Before committing:

```
make pre-commit
```

## Docker for Desktop and K3D Known Limitations

* Docker for Desktop
    * Does not enforce Kubernetes RBAC.
* K3D:
    * Requires you to import images.
    * Does not enforce resource requests.