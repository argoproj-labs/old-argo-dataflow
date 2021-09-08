from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("106-git-python")
     .owner('argoproj-labs')
     .describe("""This example of a pipeline using Git.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts. This example presents how one can use python runtime git step.

[Link to directory that is be cloned with git](../examples/git-python/)

[Learn about Git steps](../docs/GIT.md)""")
     .step(
        (kafka('input-topic')
         .git('main', 'https://github.com/argoproj-labs/argo-dataflow', 'main', 'examples/git-python',
              'quay.io/argoprojlabs/dataflow-python3-9',
              command=["/dumb-init", "--", "./start.sh"])
         .kafka('output-topic')
         ))
     .save())
