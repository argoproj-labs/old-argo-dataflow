# Workflows Interop

How to use Dataflow with Argo Workflows.

## Starting A Pipeline From A Workflow

Use a Argo Workflows resource template:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: create-pipeline-
spec:
  entrypoint: main
  templates:
    - name: main
      resource:
        action: create
        successCondition: status.phase == Succeeded
        failureCondition: status.phase in (Failed, Error)
        manifest: |
          apiVersion: dataflow.argoproj.io/v1alpha1
          kind: Pipeline
          metadata:
            name: main
          spec:
            steps:
              - cat: {}
                name: main
```

## Sending Messages From A Workflow To A Pipeline

Use a HTTP source:

```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: http
spec:
  steps:
    - cat: { }
      name: main
      sources:
        - name: http-0
          http:
            serviceName: http-svc
```

Use `curl` to send messages:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  generateName: pump-pipeline-
spec:
  entrypoint: main
  templates:
    - name: main
      container:
        command: [ curl ]
        args: [ http://http-svc/sources/http-0, -d, my-msg ]
        image: ubuntu
```

## Starting A Workflow From A Pipeline

Use the Argo Server [webhook endpoint](https://argoproj.github.io/argo-workflows/webhooks/).
Follows [this guide first](https://argoproj.github.io/argo-workflows/submit-workflow-via-automation/) to set-up an
access token, and then you can create a pipeline like this:

```yaml
apiVersion: dataflow.argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: http
spec:
  steps:
    - cat: { }
      name: main
      sinks:
        - http:
            url: https://localhost:2746/api/v1/events/my-namespace/-
            headers:
              - name: Authorization
                valueFrom:
                  # this secret should contain the access token with "Bearer " prefix
                  secretKeyRef:
                    name: my-secret-name
                    key: my-secret-key
```
