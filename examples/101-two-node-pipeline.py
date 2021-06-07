from dsls.python import kafka, pipeline, stan

if __name__ == "__main__":
    (pipeline("two-node")
     .describe("""This example shows a example of having two nodes in a pipeline.

While they read from Kafka, they are connected by a NATS Streaming subject.""")
     .step(
        (kafka('input-topic')
         .cat('a')
         .stan('a-b'))
    )
     .step(
        (stan('a-b')
         .cat('b')
         .kafka('output-topic'))
    )
     .dump())
