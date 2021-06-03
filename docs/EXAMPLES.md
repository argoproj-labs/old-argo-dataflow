### Examples

### [Hello](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/101-hello-pipeline.yaml)

This is the hello world of pipelines.

It uses a cron schedule as a source and then just cat the message to a log


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/101-hello-pipeline.yaml
```

### [Two nodes pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/101-two-node-pipeline.yaml)

This example shows a example of having two nodes in a pipeline.

While they read from Kafka, they are connected by a NATS Streaming subject.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/101-two-node-pipeline.yaml
```

### [Filter messages](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-filter-pipeline.yaml)

This is an example of built-in filtering.

Filters are written using expression syntax and must return a boolean.

They have a single variable, `msg`, which is a byte array.

[Learn about expressions](../docs/EXPRESSIONS.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-filter-pipeline.yaml
```

### [Flatten and expand](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-flatten-expand-pipeline.yaml)

This is an example of built-in flattening and expanding.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-flatten-expand-pipeline.yaml
```

### [Map messages](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-map-pipeline.yaml)

This is an example of built-in mapping.

Maps are written using expression syntax and must return a byte array.

They have a single variable, `msg`, which is a byte array.

[Learn about expressions](../docs/EXPRESSIONS.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/102-map-pipeline.yaml
```

### [Using replicas and auto-scaling](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/103-autoscaling-pipeline.yaml)

This is an example of having multiple replicas for a single step.

Replicas are automatically scaled up and down depending on the number of messages pending processing.

The ratio is defined as the number of pending messages per replica:

```
replicas = pending / ratio
```

The number of replicas will not scale beyond the min/max bounds (except when *peeking*, see below):

```
min <= replicas <= max
```

* `min` is used as the initial number of replicas.
* If `ratio` is undefined no scaling can occur; `max` is meaningless.
* If `ratio` is defined but `max` is not, the step may scale to infinity.
* If `max` and `ratio` are undefined, then the number of replicas is `min`.
* In this example, because the ratio is 1000, if 2000 messages pending, two replicas will be started.
* To prevent scaling up and down repeatedly - scale up or down occurs a maximum of once a minute.
* The same message will not be send to two different replicas.

### Scale-To-Zero and Peeking

You can scale to zero by setting `minReplicas: 0`. The number of replicas will start at zero, and periodically be scaled
to 1  so it can "peek" the the message queue. The number of pending messages is measured and the target number
of replicas re-calculated.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/103-autoscaling-pipeline.yaml
```

### [Go 1.16 handler](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-go1-16-pipeline.yaml)

This example of Go 1.16 handler.

[Learn about handlers](../docs/HANDLERS.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-go1-16-pipeline.yaml
```

### [Java 16 handler](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-java16-pipeline.yaml)

This example is of the Java 16 handler.

[Learn about handlers](../docs/HANDLERS.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-java16-pipeline.yaml
```

### [Python 3.9 handler](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-python-pipeline.yaml)

This example is of the Python 3.9 handler.

[Learn about handlers](../docs/HANDLERS.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-python-pipeline.yaml
```

### [Git handler](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/106-git-pipeline.yaml)

This example of a pipeline using Git.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts.

[Learn about Git steps](../docs/GIT.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/106-git-pipeline.yaml
```

### [Runs to completion](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/107-completion-pipeline.yaml)

This example shows a pipelne running to completion.

A pipeline that run to completion (aka "termianting") is one that will finish.

For a pipeline to terminate one of two things must happen:

* Every steps exits successfully (i.e. with exit code 0).
* One step exits successfully, and is marked with `terminator: true`. When this happens, all other steps are killed.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/107-completion-pipeline.yaml
```

### [Terminator](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/107-terminator-pipeline.yaml)

This example demostrates having a terminator step, and then terminating other steps
using different terminations strategies.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/107-terminator-pipeline.yaml
```

### [Using FIFOs for input and outputs](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/108-fifos-pipeline.yaml)

This example use named pipe to send and receive messages.

Two named pipes are made available:

* The container can read lines from `/var/run/argo-dataflow/in`. Each line will be a single message.
* The contain can write to `/var/run/argo-dataflow/out`. Each line MUST be a single message.

You MUST escape new lines.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/108-fifos-pipeline.yaml
```

### [Group messages](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/109-group-pipeline.yaml)

This is an example of built-in grouping.

There are four mandatory fields:

* `key` A string expression that returns the message's key
* `endOfGroup` A boolean expression that returns whether or not to end group and send the group messages onwards.
* `format` What format the grouped messages should be in.
* `storage` Where to store messages ready to be forwarded.

[Learn about expressions](../docs/EXPRESSIONS.md)

### Storage

Storage can either be:

* An ephemeral volume - you don't mind loosing some or all messages (e.g. development or pre-production).
* A persistent volume - you want to be to recover (e.g. production).


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/109-group-pipeline.yaml
```

### [Vetinary](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/201-vetinary-pipeline.yaml)

This pipeline processes pets (cats and dogs).


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/201-vetinary-pipeline.yaml
```

### [Word count](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/201-word-count-pipeline.yaml)

This pipeline count the number of words in a document, not the number of count of each word as you might expect.

It also shows an example of a pipelines terminates based on a single step's status.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/201-word-count-pipeline.yaml
```

### [Cron](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-cron-pipeline.yaml)

This example uses a cron source.

You can format dates using a "layout":

https://golang.org/pkg/time/#Time.Format

By deafult, the layout is RFC3339.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-cron-pipeline.yaml
```

### [Erroring](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-erroring-pipeline.yaml)

This example creates errors randomly


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-erroring-pipeline.yaml
```

### [HTTP](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-http-pipeline.yaml)

This example uses a HTTP sources and sinks.

HTTP has the advantage that it is stateless and therefore cheap. You not need to set-up any storage for your
messages between steps. Unfortunately, it is possible for some or all of your messages to not get delivered.
Also, this is sync, not async, so it can be slow due to the time taken to deliver messages.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-http-pipeline.yaml
```

### [Kafka](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-kafka-pipeline.yaml)

This example shows reading and writing to cafe.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-kafka-pipeline.yaml
```

### [Parallel](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-parallel-pipeline.yaml)

This example uses parallel to 2x the amount of data it processes.


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-parallel-pipeline.yaml
```

### [Two sinks](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-two-sinks-pipeline.yaml)

This example has two sinks


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-two-sinks-pipeline.yaml
```

### [Two sources](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-two-sources-pipeline.yaml)

This example has two sources


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/301-two-sources-pipeline.yaml
```

### [Default Kafka config](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-kafka-default-secret.yaml)

This is an example of providing a namespace named Kafka configuration.

The secret must be named `dataflow-kafka-${name}`.

# Brokers as a comma-separated list
brokers: broker.a,broker.b
# Enable TLS
net.tls: ""
# Kafka version
version: "2.0.0"

[Learn about configuration](../docs/CONFIGURATION.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-kafka-default-secret.yaml
```

### [Default NATS Streaming (STAN) configuration](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-stan-default-secret.yaml)

This is an example of providing a namespace named NATS Streaming configuration.

The secret must be named `dataflow-nats-${name}`.

[Learn about configuration](../docs/CONFIGURATION.md)


```
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-stan-default-secret.yaml
```

