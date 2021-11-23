from argo_dataflow import kafka, pipeline, stan

if __name__ == '__main__':
    (pipeline("kafka-two-step")
     .owner('acollins8')
     .step(
        (kafka('input-topic')
         .cat('a')
         .stan('a-b'))
    )
        .step(
        (stan('a-b')
         .cat('b')
         .kafka('output-topic'))
    )
        .save())
