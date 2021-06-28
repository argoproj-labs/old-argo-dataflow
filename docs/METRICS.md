# Prometheus Metrics

Each replica's sidecar exposes Prometheus metrics so you can build graphs:

```
# HELP input_inflight Number of in-flight messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_inflight
# TYPE input_inflight gauge
input_inflight{replica="0"} 0
# HELP input_message_time_seconds Message time, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_message_time_seconds
# TYPE input_message_time_seconds histogram
input_message_time_seconds_bucket{replica="0",le="0.005"} 22
input_message_time_seconds_bucket{replica="0",le="0.01"} 23
input_message_time_seconds_bucket{replica="0",le="0.025"} 23
input_message_time_seconds_bucket{replica="0",le="0.05"} 24
input_message_time_seconds_bucket{replica="0",le="0.1"} 24
input_message_time_seconds_bucket{replica="0",le="0.25"} 24
input_message_time_seconds_bucket{replica="0",le="0.5"} 24
input_message_time_seconds_bucket{replica="0",le="1"} 24
input_message_time_seconds_bucket{replica="0",le="2.5"} 24
input_message_time_seconds_bucket{replica="0",le="5"} 24
input_message_time_seconds_bucket{replica="0",le="10"} 24
input_message_time_seconds_bucket{replica="0",le="+Inf"} 24
input_message_time_seconds_sum{replica="0"} 0.052320869
input_message_time_seconds_count{replica="0"} 24
...
# HELP replicas Number of replicas, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#replicas
# TYPE replicas gauge
replicas 1
# HELP sources_errors Total number of errors, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_errors
# TYPE sources_errors counter
sources_errors{sourceName="default"} 0
# HELP sources_pending Pending messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_pending
# TYPE sources_pending counter
sources_pending{sourceName="default"} 0
# HELP sources_total Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_total
# TYPE sources_total counter
sources_total{sourceName="default"} 528
```

### input_inflight

Use this metric to determine how many message each replica can process in parallel.

Golden metric type: saturation.

### input_message_time_seconds

Use this metric to determine how long messages are taking to be processed.

Golden metric type: latency.

### replicas

Use this to track scaling events. 

Only exposed by replica 0.

Golden metric type: traffic.

### sources_errors

Use this to track errors.

Only exposed by replica 0.

Golden metric type: error.

### sources_total

Use this to track throughput. 

Only exposed by replica 0.

Golden metric type: traffic.

### sources_pending

Use this to track back-pressure.

Only exposed by replica 0.

Golden metric type: traffic.

### duplicate_messages

Use this to track duplicate messages filtered by a dedupe step.

This is exposed by the main container on port 8080, not by the sidecar or 3569.