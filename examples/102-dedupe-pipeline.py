from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("102-dedupe")
     .owner('argoproj-labs')
     .describe("""This is an example of built-in de-duplication step.""")
     .step(
        (kafka('input-topic')
         .dedupe()
         .kafka('output-topic'))
    )
        .save())
