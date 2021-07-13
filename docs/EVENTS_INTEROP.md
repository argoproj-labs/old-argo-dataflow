# Argo Events Interop

How to use Dateflow with Argo Events.

## Use EventSources As Pipeline Sources

All the EventSources from Argo Events can be used as Pipeline sources. Use a
`Calendar` EventSource as an example, it emits an event every 10 seconds:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: EventSource
metadata:
  name: calendar
spec:
  calendar:
    example:
      interval: 10s
```

Start a Pipeline like below:

```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: hello
  namespace: argo-dataflow-system
spec:
  steps:
    - map: bytes(object(msg).data)
      name: main
      sinks:
        - log: {}
      sources:
        - http:
            serviceName: http-main
```

Then use a Sensor to wire the EventSource and Pipeline:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Sensor
metadata:
  name: calendar
spec:
  dependencies:
    - name: dep
      eventSourceName: calendar
      eventName: example
  triggers:
    - template:
        name: http-trigger
        http:
          url: http://http-main/sources/default
          method: POST
          payload:
            - src:
                dependencyName: dep
                dataKey: body
              # The value of "dest" needs to be consistent with "bytes(object(msg).data)" in Pipeline definition.
              dest: data
```
