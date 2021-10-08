# Processing Big Payload >1MB
Default Dataflow configuration can process the upto 1MB payload. If you want to process bigger payload,
you need to configure the resource for `step` and `sidecar`. We have tested below large payloads on HTTP, Kafka and Stan source.

#### Resource Configuration:
```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: http
spec: 
  steps:
  - cat:
      resources:
        requests:
          memory: 1Gi
    name: main
    replicas: 5
    restartPolicy: OnFailure
    scale:
      peekDelay: defaultPeekDelay
      scalingDelay: defaultScalingDelay
    serviceAccountName: pipeline
    sidecar:
      resources:
        requests:
          memory: 1Gi
    sinks:
    - log:
        truncate: 32
      name: log
    sources:
    - http: {}
      name: default
      retry:
        cap: 0ms
        duration: 100ms
        factorPercentage: 200
        jitterPercentage: 10
        steps: 20
```



Results:
   
| Source | 1 Mb|5 Mb|10 Mb|
|---|---|---|---|
| HTTP |TPS 80 |TPS 35| TPS 17|
| Kafka | TPS 74|TBS 19|TPS 11|
|Stan|TPS 113|TPS 22|TPS 8|

| Sink | 1 Mb|5 Mb|10 Mb|
|---|---|---|---|
| HTTP |TPS 81 |TPS 34|TPS 17|
| Kafka | TPS 74|TPS 18|TPS 9|
|Stan| TPS 69| TPS 6|TPS 3|    
