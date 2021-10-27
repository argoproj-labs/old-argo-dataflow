# Kafka

If you want to experiment with Kafka, install Kafka:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/apps/kafka.yaml
```

Configure dataflow to use that Kafka by default:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-kafka-default-secret.yaml 
```

Wait for the statefulsets to be available (ctrl+c when available):

```bash
kubectl get statefulset -w
```

If you want to connect to from you desktop, e.g. as a consumer or producer, you can port forward to the Kafka broker:

```bash
kubectl port-forward svc/kafka-broker 9092:9092
```

To send messages from your host, you'll need to update `/etc/hosts`:

```
sudo vi /etc/hosts
```

Add these lines:

```
127.0.0.1 kafka-0.broker
127.0.0.1 kafka-broker
```

You can use Kafka's console producer to send messages to the broker,
see [Kafka quickstart](https://kafka.apache.org/quickstart).