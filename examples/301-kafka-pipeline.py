from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("301-kafka")
     .owner('argoproj-labs')
     .describe("""This example shows reading and writing to a Kafka topic
     
* Kafka topics are typically partitioned. Dataflow will process each partition simultaneously.
* Adding replicas will nearly linearly increase throughput.
* If you scale beyond the number of partitions, those additional replicas will be idle.
     """)
     .annotate("dataflow.argoproj.io/test", "true")
     .step(
        (kafka('input-topic', groupId='my-group')
         .cat()
         .kafka('output-topic', a_sync=True)
         ))
     .save())
