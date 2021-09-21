from argo_dataflow import pipeline, CatStep, KafkaSource, CronSource, LogSink

if __name__ == '__main__':
    (pipeline("301-two-sources")
     .owner('argoproj-labs')
     .describe("""This example has two sources
""")
     .step(
        CatStep(
            sources=[KafkaSource(
                'input-topic'), CronSource(schedule='*/3 * * * * *', layout="15:04:05")],
            sinks=[LogSink()]
        )
    )
        .save())
