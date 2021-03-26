Add to `/etc/hosts`:

```
127.0.0.1 kafka-0.broker.kafka.svc.cluster.local
```

Install the basics:

```
make kafka ;# install Kafka in cluster
```

```

```

Start pumping messages in and out of Kafka:

```
make run
```

Create an example pipeline:

```
make example
```

