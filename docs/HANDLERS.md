# Handler Step

Handlers are intended as a convenient way to write steps without having to build and publish images.

A handler is defined as:

* Code to run.
* A runtime to build and execute the code. E.g. `golang1-16`, `java16`, or `python3-9`:

## Inline


When a step starts, the code will be written to a file, and the file compiled and executed.

```yaml
handler:
  code: |
    def handler(msg, context):
      return msg
  runtime: python3-9
```

Examples:

* [Go 1.16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-golang1-16-pipeline.yaml)
* [Java 16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-java16-pipeline.yaml)
* [Python 3.9 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-python3-9-pipeline.yaml)

