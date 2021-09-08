from argo_dataflow import pipeline, git

if __name__ == '__main__':
    (pipeline("106-git-python-generator")
     .owner('argoproj-labs')
     .describe("""This example of a pipeline using Git, showing how to use Generator Step.

The Git handler allows you to check your application source code into Git. Dataflow will checkout and build
your code when the step starts. This example presents how one can use python runtime git generator step. Generator steps
are long running processes generating values over time. Such a step doesn't have a source, only sink(s).

[Link to directory that is be cloned with git](../examples/git-python-generator-step/)

[Learn about Git steps](../docs/GIT.md)""")
     .step(
        (git('main', 'https://github.com/argoproj-labs/argo-dataflow', 'main', 'examples/git-python-generator-step',
             'quay.io/argoprojlabs/dataflow-python3-9',
             command=["/dumb-init", "--", "./start.sh"])
         .kafka('output-topic')
         ))
     .save())
