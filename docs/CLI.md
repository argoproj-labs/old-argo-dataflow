# CLI

List pipelines:

```
kubectl get pipeline
```

List steps for a pipeline:

```
kubectl get step -l dataflow.argoproj.io/pipeline-name=my-pipeline
```

List pods for a pipeline step:

```
kubectl get pod -l dataflow.argoproj.io/pipeline-name=my-pipeline,step.argoproj.io/pipeline-name=my-step
```

Restart a pipeline:

```
kubectl delete pod -l dataflow.argoproj.io/pipeline-name=my-pipeline
```

Restart a step:

```
kubectl delete pod -l dataflow.argoproj.io/pipeline-name=my-pipeline,step.argoproj.io/pipeline-name=my-step
```
