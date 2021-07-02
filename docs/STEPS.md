# Steps

A step is the execution unit of a Dataflow pipeline.

## Built-in Steps

There are a number of built in steps. These are tested for reliability and performance and so should be you first
choice:

* `cat` echo messages back unchanged
* `dedupe` remove duplicate messages
* `expand` expand dot-delimited messages to structured message
* `filter` filter messages
* `flatten` flatten structured message to dot-delimited messages
* `map` map messages to new messages

## Code Steps

These are two step that you can specify code:

* `git` define a place in Git to check the code out from and run
* `handler` define the code inline

## Container Step

A `container` step allows you to specify the container images to run. This is fully flexible and powerful. The image
must obey [the image contract](IMAGE_CONTRACT.md).

Both built-in steps, and code steps, could be re-written as container steps (if fact, that is what they all are
underneath). underneath)

### Generator Step

This is a step that sinks data using the `:3569/messages` endpoint. This allows you to create a step that puts data from
a source dataflow does not yet support (e.g. AWS Kinesis) into a place that it does support (e.g. Kafka).

It would be weird to have any sources for this step, and you'd typically ignore any messages on the `:8080/messages`
endpoint.

### Message Hole Step

This is a step that never returns messages to sink. This allows you to create a step that puts data into a sink dataflow
does not yet support (e.g. AWS Kinesis) from a place that it does support (e.g. Kafka).

You would not typically have any sink on this step, because it would never get anything to sink.