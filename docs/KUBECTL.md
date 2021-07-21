# `kubctl`

Argo Dataflow is designed to work well with `kubectl`, rather than needing its own CLI (though maybe we'll add one
someday 😀).

Task you can do with `kubectl`:

Create a pipeline:

```
kubectl apply -f examples/101-hello-pipeline.yaml
```

Wait for the pipeline to be running:

```
kubectl wait pipeline/101-hello --for=condition=running
```

Restart pipeline:

```
kubectl delete pod -l dataflow.argoproj.io/pipeline-name=xxx
```