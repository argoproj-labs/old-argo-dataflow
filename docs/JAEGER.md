# Jaeger

We use conventional configuration for Jaeger, example

```
export JAEGER_DISABLED=false
export JAEGER_ENDPOINT=http://my-jaeger-collector:14268/api/traces
# export JAEGER_REPORTER_LOG_SPANS=true
# sample one message per second
export JAEGER_SAMPLER_TYPE=ratelimiting
export JAEGER_SAMPLER_PARAM=0.2
```