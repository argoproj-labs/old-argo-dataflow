from argo_dataflow import pipeline, stan

if __name__ == '__main__':
    (pipeline("301-stan")
     .owner('argoproj-labs')
     .describe("""This example shows reading and writing to a STAN subject.

* Adding replicas will nearly linearly increase throughput.       
""")
     .annotate('dataflow.argoproj.io/test', 'false')
     .step(
        (stan('input-subject')
         .cat('main')
         .stan('output-subject')
         ))
     .save())
