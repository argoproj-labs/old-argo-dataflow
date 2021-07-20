from argo_dataflow import pipeline, cron


def handler(msg, context):
    import random
    if random.randint(0, 4) == 1:
        raise Exception("random error")
    return msg


if __name__ == '__main__':
    (pipeline("301-erroring")
     .owner('argoproj-labs')
     .describe("""This example showcases retry policies.""")
     .annotate('dataflow.argoproj.io/wait-for', 'RecentErrors')
     .step(
        (cron('*/3 * * * * *', retry={'steps': 99999999})
         .code('always', source=handler)
         .log())
    )
     .step(
        (cron('*/3 * * * * *', retry={'steps': 0})
         .code('never', source=handler)
         .log())
    ).save())
