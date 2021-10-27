from argo_dataflow import pipeline, kafka


def a(msg, context):
    return ("hi! " + msg.decode("UTF-8")).encode("UTF-8")


def b(msg, context):
    return ("bye! " + msg.decode("UTF-8")).encode("UTF-8")


if __name__ == '__main__':
    (pipeline("104-python3-9")
     .owner('argoproj-labs')
     .describe("""This example is of the Python 3.9 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .step(
        (kafka('input-topic')
         .code('a', source=a)
         .kafka('middle-topic'))
    )
        .step(
        (kafka('middle-topic')
         .code('b', source=b)
         .kafka('output-topic'))
    )
        .save())
