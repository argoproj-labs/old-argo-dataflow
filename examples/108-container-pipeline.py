from argo_dataflow import pipeline, container

if __name__ == '__main__':
    (pipeline("108-container")
     .owner('argoproj-labs')
     .describe("""This example showcases container options.""")
     .annotate('dataflow.argoproj.io/wait-for', 'Completed')
     .step(
        (container('main',
                   args=['sh', '-c', 'exit 0'],
                   image='ubuntu:latest',
                   env={'FOO': 'bar'},
                   resources={'requests': {'cpu': 1}}
                   )
         .annotations({'my-annotation': 'my-value'})
         ))
     .save())
