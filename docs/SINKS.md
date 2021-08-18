# Sinks

If a message cannot be sunk, it will error immediately, and the message fail completely. This error bubbles up to to the
source, and therefore will be retries as per the source's configuration.

## Database

Consumes messages from a database by periodically running SQL queries.

## HTTP

Makes a HTTP request.

[Example](../examples/301-http-pipeline.py)

## Log

Logs the message.

[Example](../examples/301-cron-log-pipeline.py)

## Kafka

Writes messages to a Kafka topic.

[Example](../examples/301-kafka-pipeline.py)

## NATS Streaming (STAN)

Writes messages to a NATS streaming subject.

[Example](../examples/301-stan-pipeline.py)

## S3

Writes files to a S3 bucket.

## Volume

Writes files to a volume:

* [Storage Volumes](https://kubernetes.io/docs/concepts/storage/volumes/) e.g. NFS, Azure File, Config Map, Secret
* [Container Storage Interface (CSI) Drivers](https://kubernetes-csi.github.io/docs/drivers.html) e.g. AWS EBS, Google
  Cloud Storage
* [S3](https://github.com/ctrox/csi-s3) (not production ready)

This is an esoteric sink. The container writes a file in `/var/run/argo-dataflow/sinks/{sinkName}/`, and return an empty
message. 