# Quick Start

Deploy:

```
make deploy CONFIG=quick-start
```



Access the user interface:

```
kubectl port-forward svc/argo-server 2746:2746
```

Open [http://localhost:2746/pipelines](http://localhost:2746/pipelines).

Debugging: 

```
make logs
```

Clean up:

```
make undeploy
```