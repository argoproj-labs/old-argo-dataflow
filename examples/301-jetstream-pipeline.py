from argo_dataflow import pipeline, jetstream

if __name__ == '__main__':
    (pipeline("301-jetstream")
     .owner('argoproj-labs')
     .describe("""This example shows reading and writing to a JetStream subject.

* Adding replicas will nearly linearly increase throughput.       
""")
     .annotate('dataflow.argoproj.io/test', 'false')
     .step(
        (jetstream('input-subject')
         .cat()
         .jetstream('output-subject')
         ))
     .save())
