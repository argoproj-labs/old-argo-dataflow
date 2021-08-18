# Sources

## Cron

A cron source creates a message containing the cron schedule time with a layout.

[Example](../examples/301-cron-log-pipeline.py)

## Database

Periodically queries a database for messages.

## HTTP

Exposes a HTTP service.

[Example](../examples/301-http-pipeline.py)

## Kafka

Consumes messages from a Kafka topic.

[Example](../examples/301-kafka-pipeline.py)

## NATS Streaming (STAN)

Consumes messages from a NATS streaming subject.

[Example](../examples/301-stan-pipeline.py)

## Volume

Periodically queries a volume for files to process.