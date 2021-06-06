# Prometheus Metrics

Each replica's sidecar exposes Prometheus metrics so you can build graphs:

```
# HELP input_inflight Number of in-flight message
# TYPE input_inflight gauge
input_inflight{replica="0"} 2
# HELP sources_errors Total number of errors
# TYPE sources_errors counter
sources_errors{replica="0",sourceName="default"} 0
# HELP sources_pending Pending messages
# TYPE sources_pending counter
sources_pending{sourceName="default"} 48694
# HELP sources_total Total number of messages
# TYPE sources_total counter
sources_total{replica="0",sourceName="default"} 7771
```

## input_inflight

Use this metric to determine how many message a pod can process in parallel.

Golden metric type: traffic and saturation.

## sources_errors

Use this to track errors.

Golden metric type: error.

## sources_total

Use this to track throughput.

Golden metric type: traffic.

## sources_pending

Use this to track back-pressure.

Golden metric type: traffic.