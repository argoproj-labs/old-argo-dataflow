from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("106-git")
     .owner('argoproj-labs')
     .describe("""This example of a pipeline using Git.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts.

[Learn about Git steps](../docs/GIT.md)""")
     .step(
        (kafka('input-topic')
         .git('main', 'https://github.com/argoproj-labs/argo-dataflow', 'main', 'examples/git', 'quay.io/argoproj/dataflow-golang1-16:latest')
         .kafka('output-topic')
         ))
     .save())
