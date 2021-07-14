# Sinks

If a message cannot be sunk, it will error immediately, and the message fail completely. This error bubbles up to to the
source, and therefore will be retries as per the source's configuration.

## HTTP

Makes a HTTP request.

[Example](../examples/301-http-pipeline.py)

## Log

Logs the message.

[Example](../examples/301-cron-log-pipeline.py)

## Kafka

Consumes messages from a Kafka topic.

[Example](../examples/301-kafka-pipeline.py)

## NATS Streaming (STAN)

Consumes messages from a NATS streaming subject.

[Example](../examples/301-stan-pipeline.py)

