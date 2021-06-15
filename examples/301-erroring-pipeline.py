from argo_dataflow import pipeline, cron


def handler(msg):
    import random
    if random.randint(0, 1) == 1:
        raise Exception("random error")
    return msg


if __name__ == '__main__':
    (pipeline("301-erroring")
     .owner('argoproj-labs')
     .describe("""This example showcases retry policies.""")
     .annotate('dataflow.argoproj.io/wait-for', 'RecentErrors')
     .step(
        (cron('*/3 * * * * *', retryPolicy='Always')
         .handler('always', handler=handler)
         .log())
    )
     .step(
        (cron('*/3 * * * * *', retryPolicy='Never')
         .handler('never', handler=handler)
         .log())
    ).save())
