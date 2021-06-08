from dsls.python import pipeline, kafka

if __name__ == '__main__':
    (pipeline("301-kafka")
     .owner('argoproj-labs')
     .describe("""This example shows reading and writing to a Kafka topic""")
     .annotate("dataflow.argoproj.io/test", "true")
     .step(
        (kafka('input-topic')
         .cat('main')
         .kafka('output-topic')
         ))
     .save())
