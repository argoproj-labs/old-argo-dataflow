# Quick Start

Deploy into the `argo-dataflow-system` namespace:

```bash
kubectl create ns argo-dataflow-system
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/quick-start.yaml
```

If you want to experiment with Kafka or NATS Streaming (aka STAN):

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/kafka-dev.yaml 
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-kafka-default.yaml 
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/stan-dev.yaml 
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-stan-default.yaml
```

Change to the installation namespace:

```bash
kubectl config set-context --current --namespace=argo-dataflow-system
```

Wait for the deployments to be ready:

```bash
kubectl get deploy -w
```

Access the user interface:

```bash
kubectl port-forward svc/argo-server 2746:2746
```

Open [http://localhost:2746/pipelines/argo-dataflow-system](http://localhost:2746/pipelines/argo-dataflow-system).

Run [one of the examples](EXAMPLES.md).

Clean up:

```bash
kubectl delete ns argo-dataflow-system
```