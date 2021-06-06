# Prometheus Metrics

Each replica's sidecar exposes Prometheus metrics so you can build graphs:

```
# HELP input_inflight Number of in-flight messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_inflight
# TYPE input_inflight gauge
input_inflight{replica="0"} 0
# HELP input_message_time_seconds Message time, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#input_message_time_seconds
# TYPE input_message_time_seconds histogram
input_message_time_seconds_bucket{replica="0",le="0.005"} 33
input_message_time_seconds_bucket{replica="0",le="0.01"} 33
input_message_time_seconds_bucket{replica="0",le="0.025"} 33
input_message_time_seconds_bucket{replica="0",le="0.05"} 33
input_message_time_seconds_bucket{replica="0",le="0.1"} 33
input_message_time_seconds_bucket{replica="0",le="0.25"} 33
input_message_time_seconds_bucket{replica="0",le="0.5"} 33
input_message_time_seconds_bucket{replica="0",le="1"} 33
input_message_time_seconds_bucket{replica="0",le="2.5"} 33
input_message_time_seconds_bucket{replica="0",le="5"} 33
input_message_time_seconds_bucket{replica="0",le="10"} 33
input_message_time_seconds_bucket{replica="0",le="+Inf"} 33
input_message_time_seconds_sum{replica="0"} 0.025579325999999996
input_message_time_seconds_count{replica="0"} 33
# HELP sources_errors Total number of errors, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_errors
# TYPE sources_errors counter
sources_errors{replica="0",sourceName="default"} 0
# HELP sources_pending Pending messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_pending
# TYPE sources_pending counter
sources_pending{sourceName="default"} 0
# HELP sources_total Total number of messages, see https://github.com/argoproj-labs/argo-dataflow/blob/main/docs/METRICS.md#sources_total
# TYPE sources_total counter
sources_total{replica="0",sourceName="default"} 33
```

## input_inflight

Use this metric to determine how many message a pod can process in parallel.

Golden metric type: saturation.

## input_message_time_seconds

Use this metric to determine how long messages are taking to be processed.

Golden metric type: latency.

## sources_errors

Use this to track errors.

Golden metric type: error.

## sources_total

Use this to track throughput.

Golden metric type: traffic.

## sources_pending

Use this to track back-pressure.

Golden metric type: traffic.