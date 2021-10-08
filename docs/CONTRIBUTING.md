# Contributing

Install :

* Golang v1.16
* Docker
* Kubernetes cluster ready (e.g. Docker for Desktop).
* Python 3 and Pip 3
* [protoc](https://grpc.io/docs/protoc-installation/).

Start:

    make start

To start the user interface, you must have checked out Argo Workflows in `../../argoproj/argo-workflows`. The UI will
appear on port 8080, then run `make start UI=true`.

If need to run code generation:

    make codegen

Before you commit, run:

    make pre-commit -B

Then:

* Sign-off your commits.
* Use [Conventional Commit messages](https://www.conventionalcommits.org/en/v1.0.0/).
* Suffix the issue number.

Example:

    git commit --signoff -m 'fix: Fixed broken thing. Fixes #1234'

## Docker for Desktop and K3D Known Limitations

* Docker for Desktop
    * Does not enforce Kubernetes RBAC.
* K3D:
    * Requires you to import images.
    * Does not enforce resource requests.
