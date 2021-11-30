# Concepts

## Pipelines & Steps

A **pipeline** is specified by a Kubernetes custom resource and consists of one or more **steps**. Each step is
specified as a
**processor**, one or more **sources**, and one or more **sinks**. Messages are read from the source, sent to the
processor, and then maybe written the output to the sink.

## Sources

A source is somewhere to get messages from, e.g.:

* HTTP - a HTTP endpoint, the HTTP body is the message,
* Kafka - a Kafka topic
* Jetstream - a Jetstream subject

[Learn more](SOURCES.md)

## Processors

A processor is a function that processes messages:

* Built-in:
    * Filter - filter out messages based on an expression
    * Map - map messages to new messages
* Code - run Golang or Python function
* Git - checkout a function from git and run it.
* Container - run a container image to process the function

All processors are actually containers that just listen on http://localhost:8080/messages and messages are sent in the
body of each request to that endpoint. If the endpoint returns a body, then it is sent to the sink.

You can write a processor in any language that can receive HTTP requests, and we provide SDKs so you don't even need to
write that. Dataflow takes care of the messaging for you.

[Learn more](PROCESSORS.md)

## Sinks

A sink is somewhere a message can be sent to, e.g.:

* Log - write the message to a log file
* HTTP - make a HTTP request where the message is the body
* Kafka - write to a Kafka topic

[Learn more](SINKS.md)
