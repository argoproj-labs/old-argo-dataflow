from argo_dataflow import cron, pipeline

if __name__ == '__main__':
    (pipeline("101-hello")
     .owner('argoproj-labs')
     .namespace('argo-dataflow-system')
     .describe("""This is the hello world of pipelines.

It uses a cron schedule as a source and then just cat the message to a log""")
     .annotate('dataflow.argoproj.io/test', "true")
     .step(
        (cron('*/3 * * * * *')
         .cat('main')
         .log())
    )
     .save())
