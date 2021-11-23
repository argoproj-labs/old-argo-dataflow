# Code Step

Handlers are intended as a convenient way to write steps without having to build and publish images.

A handler is defined as:

* Code to run.
* A runtime to build and execute the code. E.g. `golang1-17`, `java16`, `python3-9` or `node16`:

## Inline

When a step starts, the code will be written to a file, and the file compiled and executed.

```yaml
code:
  source: |
    def handler(msg, context):
      return msg
  runtime: python3-9
```

Examples:

* [Go 1.17 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-golang1-17-pipeline.yaml)
* [Java 16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-java16-pipeline.yaml)
* [Python 3.9 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-python3-9-pipeline.yaml)
* [NodeJS 16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-node16-pipeline.yaml)

