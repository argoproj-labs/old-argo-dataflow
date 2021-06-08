from dsls.python import pipeline, container

if __name__ == '__main__':
    (pipeline("108-container")
     .describe("""This example showcases container options.""")
     .step(
        (container('main',
                   args=['sh', '-c', 'exit 0'],
                   image='ubuntu:latest',
                   env={'FOO': 'bar'},
                   )
         .annotations({'my-annotation': 'my-value'})
         ))
     .save())
