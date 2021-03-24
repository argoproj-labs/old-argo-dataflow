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
```

Made a change?

```
kubectl rollout restart deploy controller-manager
```

```
kubectl delete pipeline --all
kubectl apply -f example-pipeline.yaml
```