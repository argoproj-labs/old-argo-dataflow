from dsls.python import pipeline, kafka

if __name__ == '__main__':
    (pipeline("parallel")
     .describe("""This example uses parallel to 2x the amount of data it processes.""")
     .annotate("dataflow.argoproj.io/test", "false")
     .step(
        (kafka('input-topic', parallel=2)
         .cat('main')
         .log()
         ))
     .dump())
