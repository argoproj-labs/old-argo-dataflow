# Contributing

Add to `/etc/hosts`:

```
127.0.0.1 kafka-0.broker.kafka.svc.cluster.local
```

Start it:

```
make start
```

Create an example pipeline:

```
make examples/http-pipeline.yaml
```

