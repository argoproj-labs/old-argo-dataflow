# Quick Start

You'll need a Kubernetes cluster to try it out, e.g. Docker for Desktop.

Deploy into the `argo-dataflow-system` namespace:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/quick-start.yaml
```

Change to the installation namespace:

```bash
kubectl config set-context --current --namespace=argo-dataflow-system
```

Wait for the deployments to be available (ctrl+c when available):

```bash
kubectl get deploy -w
```

If you want to experiment with Kafka:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/apps/kafka.yaml
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-kafka-default-secret.yaml 
```

Wait for the statefulsets to be available (ctrl+c when available):

```bash
kubectl get statefulset -w
```

You can port forward to the Kafka broker:

```bash
kubectl port-forward svc/kafka-broker 9092:9092
```

You can use Kafka's console producer to send messages to the broker,
see [Kafka quickstart](https://kafka.apache.org/quickstart).

If you want the user interface:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/apps/argo-server.yaml
kubectl get deploy -w ;# (ctrl+c when available)
kubectl port-forward svc/argo-server 2746:2746
```

Open [http://localhost:2746/pipelines/argo-dataflow-system](http://localhost:2746/pipelines/argo-dataflow-system).

Run [one of the examples](EXAMPLES.md).
