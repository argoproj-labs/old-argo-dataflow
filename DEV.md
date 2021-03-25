Add to `/etc/hosts`:

```
127.0.0.1 kafka-0.broker.kafka.svc.cluster.local
```

Install the basics:

```
make kafka ;# start Kafka in cluster
```

```
go install github.com/Shopify/sarama/tools/kafka-console-consumer
go install github.com/Shopify/sarama/tools/kafka-console-producer

go run ./topic-creator -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic my-topic
go run ./topic-creator -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic your-topic
```

Start pumping messages in and out of Kafka:

```
kafka-console-consumer -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic your-topic
```

```
while true ; do kafka-console-producer -brokers kafka-0.broker.kafka.svc.cluster.local:9092 -topic my-topic -value my-val ; sleep 1; done 
```

Deploy dataflow:

```
make deploy 
```

Create an example pipeline:

```
make example
```

Made a change?

```
make redeploy logs
```
