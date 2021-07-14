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

Start a Pipeline like below, which uses a `log` sink for demostration. This
pipeline contains 2 steps, first step filter the events from defined
EventSource, second step decode the message and print it in the log.

```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: events-pipeline
spec:
  steps:
    # Use EventSourceName as source, EventName as subject.
    # Reference to https://github.com/argoproj/argo-events/blob/master/docs/eventsources/naming.md for EventSource naming detail.
    - filter: |-
        object(msg).source == "calendar" && object(msg).subject == "example"
      name: filter
      sources:
        - stan:
            name: eventbus
            natsUrl: nats://eventbus-default-stan-svc:4222
            natsMonitoringUrl: http://eventbus-default-stan-svc:8222
            clusterId: eventbus-default
            subject: eventbus-argo-dataflow-system
      sinks:
        - stan:
            subject: filtered
    - map: bytes(sprig.b64dec(object(msg).data_base64))
      name: main
      sources:
        - stan:
            subject: filtered
      sinks:
        - log: {}
```

`stan` config in the filter can be found by viewing EventBus status
`kubectl get eb eventbus-name -o yaml`. An example status looks like below:

```yaml
status:
  config:
    nats:
      auth: none
      clusterID: eventbus-default
      url: nats://eventbus-default-stan-svc:4222
      monitoringUrl: http://eventbus-default-stan-svc:8222
```

- `subject` in `stan` config is `eventbus-{namespace}`.
