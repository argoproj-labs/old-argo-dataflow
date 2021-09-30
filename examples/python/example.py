from argo_dataflow import cron, pipeline

if __name__ == '__main__':
    (pipeline('hello')
     .namespace('argo-dataflow-system')
     .step(
        (cron('*/3 * * * * *')
         .cat()
         .log())
    )
        .run())
