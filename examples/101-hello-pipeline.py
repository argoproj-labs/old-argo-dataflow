from dsls.python import cron, pipeline

if __name__ == '__main__':
    (pipeline("hello")
     .describe("""This is the hello world of pipelines.

It uses a cron schedule as a source and then just cat the message to a log""")
     .annotate('dataflow.argoproj.io/test', "true")
     .step(
        (cron('*/3 * * * * *')
         .cat('main')
         .log())
    )
     .dump())
