# Dataflow vs...

A **pipeline** typically runs forever processing an infinite stream of small (<1MB) items of **data**.

Use dataflow:

* If you're processing a large number (could be infinite) of data items.
* If you don't know how many items of data you'll be processing.
* If you're processing a lot of small messages.
* If your data is in Kafka, NATS streaming, S3, a database, or can be sent via HTTP.
* If you're doing data processing.

## ...Argo Workflows

A **workflow** typically runs a fixed number (<10k) of **tasks**, where each task processes one large file (>1MB).

Use workflows:

* If you're executing 100s or 1000s tasks rather than processing data.
* If your data is stored as files in a bucket.
* If you're processing some large files.
* If your data is in S3/HDFS/Git or other file/bucket storage.
* If you're doing batch processing.

## ...Argo Events

An **event source** runs forever processing an infinite stream of **events**, and is connected to a **sensor** which
triggers a workflows for each event.

Use events:

* If you want to trigger actions based on events.
* If the events come in via a supported events source (including things like SQS, Apache Pulsar, Slack).
* If you're processing a small number of events.
* If you're doing event processing.
