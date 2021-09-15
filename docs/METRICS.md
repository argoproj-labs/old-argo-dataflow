# Prometheus Metrics

All metrics are approximate.

## Sidecar Metrics

The lead replica's (replica 0) sidecar exposes Prometheus metrics so you can build graphs and monitoring.

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

### sinks_total

Use this to track throughput. Includes retries and errors.

### sources_errors

Use this to track errors.

Only exposed by replica 0.

Golden metric type: error.

### sources_retries

Use this metric to determine how many retries performed for message processing.

Golden metric type: error.

### sources_total

Use this to track throughput.

Only exposed by replica 0.

Golden metric type: traffic.

### sources_pending

Use this to track back-pressure.

Only exposed by replica 0.

Golden metric type: traffic.

## Main Container Metrics

You may expose Prometheus endpoint on the main container if you want. There is nothing special about this.

### duplicate_messages

Use this to track duplicate messages filtered by a dedupe step.

This is exposed by the main container on port 8080, not by the sidecar or 3569.

