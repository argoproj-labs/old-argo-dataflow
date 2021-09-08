from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("106-git-go")
     .owner('argoproj-labs')
     .describe("""This example of a pipeline using Git.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts. This example presents how one can use go runtime git step.

[Link to directory that is be cloned with git](../examples/git/)

[Learn about Git steps](../docs/GIT.md)""")
     .step(
        (kafka('input-topic')
         .git('main', 'https://github.com/argoproj-labs/argo-dataflow', 'main', 'examples/git', 'golang:1.16',
              env=[{'name': "GOCACHE", 'value': "/tmp/.gocache"}], command=["go", "run", "."])
         .kafka('output-topic')
         ))
     .save())
