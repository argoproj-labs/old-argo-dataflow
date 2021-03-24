Add to `/etc/hosts`:

```
127.0.0.1 kafka-0.broker.kafka.svc.cluster.local
```

Install the basics:

```
make kafka
make docker-build
make deploy
kubens argo-dataflow-system
go install github.com/Shopify/sarama/tools/kafka-console-consumer
go install github.com/Shopify/sarama/tools/kafka-console-producer
export KAFKA_PEERS=kafka-0.broker.kafka.svc.cluster.local:9092
go run ./create-topic -topic my-topic
go run ./create-topic -topic your-topic
```

Made a change?

```
kubectl rollout restart deploy controller-manager
```

```
kubectl delete pipeline --all
kubectl apply -f example-pipeline.yaml
stern . -l dataflow.argoproj.io/pipeline-name
```

Send some data though the system:

```
kafka-console-consumer -topic your-topic
while true ; do kafka-console-producer -topic my-topic -value my-val ; sleep 1; done
```