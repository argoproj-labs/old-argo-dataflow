from dsls.python import pipeline, stan

if __name__ == '__main__':
    (pipeline("301-parallel")
     .owner('argoproj-labs')
     .describe("""This example uses parallel to 2x the amount of data it processes.""")
     .annotate("dataflow.argoproj.io/test", "false")
     .step(
        (stan('input-subject', parallel=2)
         .cat('main')
         .log()
         ))
     .save())
