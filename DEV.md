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
go run ./create-topic -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic my-topic
go run ./create-topic -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic your-topic
```

Made a change?

```
kubectl rollout restart deploy controller-manager
```

```
kubectl delete pipeline --all
kubectl apply -f example-pipeline.yaml
```


```
kafka-console-consumer -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic your-topic
while true ; do kafka-console-producer -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic my-topic -value my-val ; sleep 1; done
```