# Argo Dataflow

[![Build](https://github.com/argoproj-labs/argo-dataflow/actions/workflows/build.yml/badge.svg)](https://github.com/argoproj-labs/argo-dataflow/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/argoproj-labs/argo-dataflow/branch/main/graph/badge.svg?token=yKtOCXJu1Q)](https://codecov.io/gh/argoproj-labs/argo-dataflow)

## Summary

Argo Dataflow is intended as a cloud-native and language-agnostic platform for executing large parallel data-processing
pipelines composed of many steps which are often small and homogenic.

## Use Cases

* Real-time "click" analytics
* Anomaly detection
* Fraud detection
* Operational (including IoT) analytics

## Screenshot

![Screenshot](docs/assets/screenshot.png)

## Example

```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: example
  annotations:
    dataflow.argoproj.io/name: "Example pipeline"
spec:
  steps:
    - name: find-cats
      sources:
        - kafka:
            topic: pets
      filter: 'object(msg).type == "cat"'
      sinks:
        - stan:
            subject: cats

    - name: hi-cats
      sources:
        - stan:
            subject: cats
      map: '"hello " + object(msg).name'
      replicas:
        min: 2
      sinks:
        - kafka:
            topic: hello-to-cats
```

## Documentation

* [Examples](docs/EXAMPLES.md)
* [Configuration](docs/CONFIGURATION.md)
* [Handlers](docs/HANDLERS.md)
* [Git usage](docs/GIT.md)
* [Command line](docs/CLI.md)
* [Expression syntax](docs/EXPRESSIONS.md)
* [Reading material](docs/READING.md)
* [Contributing](docs/CONTRIBUTING.md)

### Architecture Diagram

[![Architecture](docs/assets/architecture.png)](https://docs.google.com/drawings/d/1Dk7mgZ3jKpBg_DQ3c8og04ULoKpGTGUt52pBE-Vet2o/edit)
