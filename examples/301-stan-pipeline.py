from dsls.python import pipeline, stan

if __name__ == '__main__':
    (pipeline("301-stan")
     .owner('argoproj-labs')
     .describe("""This example shows reading and writing to a STAN subject""")
     .annotate('dataflow.argoproj.io/test', 'false')
     .step(
        (stan('input-subject')
         .cat('main')
         .stan('output-subject')
         ))
     .save())
