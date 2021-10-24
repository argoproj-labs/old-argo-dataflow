from argo_dataflow import pipeline, container

if __name__ == '__main__':
    (pipeline("107-completion")
     .owner('argoproj-labs')
     .describe("""This example shows a pipeline running to completion.

A pipeline that run to completion (aka "terminating") is one that will finish.

For a pipeline to terminate one of two things must happen:

* Every steps exits successfully (i.e. with exit code 0).
* One step exits successfully, and is marked with `terminator: true`. When this happens, all other steps are killed.""")
     .annotate('dataflow.argoproj.io/wait-for', 'Completed')
     .step(
        (container('main',
                   args=['sh', '-c', 'exit 0'],
                   image='golang:1.17'
                   )
         ))
     .save())
