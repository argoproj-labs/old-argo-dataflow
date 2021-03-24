Add to `/etc/hosts`:

```
127.0.0.1 kafka-0.broker.kafka.svc.cluster.local
```

Install the basics:

```
make kafka ;# start Kafka in cluster
```

```
kubectl apply -f https://raw.githubusercontent.com/argoproj/argo-events/master/manifests/base/crds/argoproj.io_eventbus.yaml
```

```
go install github.com/Shopify/sarama/tools/kafka-console-consumer
go install github.com/Shopify/sarama/tools/kafka-console-producer

export KAFKA_PEERS=kafka-0.broker.kafka.svc.cluster.local:9092 ;# used by tools

go run ./topic-creator -topic my-topic
go run ./topic-creator -topic your-topic
```

Start pumping messages in and out of Kafka:

```
kafka-console-consumer -topic your-topic
```

```
while true ; do kafka-console-producer -topic my-topic -value my-val ; sleep 1; done 
```

Deploy dataflow:

```
make deploy 
```

Create a pipeline:

```
kubectl -n argo-dataflow-system delete pipeline --all
kubectl -n argo-dataflow-system apply -f example-pipeline.yaml
```

Made a change?

```
make redeploy logs
```
