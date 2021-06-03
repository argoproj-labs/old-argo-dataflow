# Handler Step

Handlers are intended as a convenient way to write steps without having to build and publish images.

A handler is defined as:

* Code to run.
* A runtime to build and execute the code. E.g. `go1-16` or `java16`.

## Inline

When a step starts, the code will be written to a file, and the file compiled and executed.

```yaml
handler:
  code: |
    package main

    func Handler(m []byte) ([]byte, error) {
      return []byte("hello " + string(m)), nil
    }
  runtime: go1.16
```

The code must include a handler function, in pseudo code:

```
Handler(msg []byte) -> ([]byte, error) 
```

Examples:

* [Go 1.16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-go1-16-pipeline.yaml)
* [Java 16 pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/104-java16-pipeline.yaml)

