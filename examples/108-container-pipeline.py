from dsls.python import pipeline, container

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
                   )
         .annotations({'my-annotation': 'my-value'})
         ))
     .save())
