from dsls.python import pipeline, kafka


def handler(msg):
    return msg


if __name__ == '__main__':
    (pipeline("104-python3-9")
     .describe("""This example is of the Python 3.9 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .annotate('dataflow.argoproj.io/timeout', '2m')
     .step(
        (kafka('input-topic')
         .handler('main', handler)
         .kafka('output-topic')
         ))
     .save())
