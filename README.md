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

```python
from dsls.python import cron, pipeline

if __name__ == '__main__':
    (pipeline('hello')
     .step(
        (cron('*/3 * * * * *')
         .cat('main')
         .log())
    )
     .dump())
```

## Documentation

* [Quick start](docs/QUICK_START.md)
* [Examples](docs/EXAMPLES.md)
* [Configuration](docs/CONFIGURATION.md)
* [Handlers](docs/HANDLERS.md)
* [Git usage](docs/GIT.md)
* [Command line](docs/CLI.md)
* [Expression syntax](docs/EXPRESSIONS.md)
* [Metrics](docs/METRICS.md)
* [Image Contract](docs/IMAGE_CONTRACT.md)
* [Reading material](docs/READING.md)
* [Contributing](docs/CONTRIBUTING.md)

### Architecture Diagram

[![Architecture](docs/assets/architecture.png)](https://docs.google.com/drawings/d/1Dk7mgZ3jKpBg_DQ3c8og04ULoKpGTGUt52pBE-Vet2o/edit)
