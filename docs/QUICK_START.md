# Quick Start

Deploy into the `argo-dataflow-system` namespace:

```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/quick-start.yaml
```

Change to the installation namespace:

```
kubectl config set-context --current --namespace=argo-dataflow-system
```

Access the user interface:

```
kubectl port-forward svc/argo-server 2746:2746
```

Open [http://localhost:2746/pipelines](http://localhost:2746/pipelines).

Run [one of the examples](EXAMPLES.md).

Clean up:

```
kubectl delete ns argo-dataflow-system
```