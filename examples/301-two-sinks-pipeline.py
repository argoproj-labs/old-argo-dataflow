from argo_dataflow import pipeline, CatStep, KafkaSource, LogSink

if __name__ == '__main__':
    (pipeline("301-two-sinks")
     .owner('argoproj-labs')
     .describe("""This example has two sinks.

* When using two sinks, you should put the most reliable first in the list, if the message cannot be delivered,
then subsequent sinks will get the message.
""")
     .step(
        CatStep(
            sources=[KafkaSource('input-topic')],
            sinks=[LogSink('a'), LogSink('b')]
        )
    )
        .save())
