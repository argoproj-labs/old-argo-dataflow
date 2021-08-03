from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("106-git-nodejs")
     .owner('argoproj-labs')
     .describe("""This example of a pipeline using Git with NodeJS.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts. This example presents how one can use nodejs runtime git step.

[Learn about Git steps](../docs/GIT.md)""")
     .step(
        (kafka('input-topic')
         .git('main', 'https://github.com/argoproj-labs/argo-dataflow', 'main', 'examples/git-nodejs', 'node:16', command=["./start.sh"])
         .kafka('output-topic')
         ))
     .save())
