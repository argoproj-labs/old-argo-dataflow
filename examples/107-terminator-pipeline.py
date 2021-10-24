from argo_dataflow import pipeline, container

if __name__ == '__main__':
    (pipeline("107-terminator")
     .owner('argoproj-labs')
     .describe("""This example demonstrates having a terminator step, and then terminating other steps
      using different terminations strategies.""")
     .annotate('dataflow.argoproj.io/wait-for', 'Completed')
     .step(
        (container('main',
                   args=['sh', '-c', 'cat'],
                   image='golang:1.17'
                   )
         ))
     .step(
        (container('terminator',
                   args=['sh', '-c', 'exit 0'],
                   image='golang:1.17',
                   terminator=True
                   )
         ))
     .save())
