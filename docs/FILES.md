# Files

Initially, Dataflow was aimed at processing messages from streaming sources such as Kafka or NATS Streaming. However, it
has quickly become clear that there are other types of data users want to process, but based on some other core
concept (files, database record).

Files are not really suitable to bundling into a message, as a file could be 1GB and messages size should be smaller (
10KB, maybe 64KB), it's too much data to keep in memory.

Instead, when we work with files, we use a shared volume and FIFOs. E.g. when an S3 source gets a message, it'll create
a FIFO at `/var/run/argo-dataflow/sources/default/the-file`, and use the S3 API `GetObject` to write data to this FIFO
for consumption by the main container.
