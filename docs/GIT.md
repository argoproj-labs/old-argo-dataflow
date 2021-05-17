# Git Step

This intended as a convenient way to write steps without having to build and publish images.

When a steps starts, the code is checked out from Git, and then run:

```yaml
git:
  branch: main
  path: examples/git
  image: golang:1.16
  url: https://github.com/argoproj-labs/argo-dataflow
  command:
    - sh
    - -c
    - |
      go run .
```

* [Example pipeline](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/106-git-pipeline.yaml)
* [Source code](https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/git)

To use this type of step, you'll need to ensure your pods can read from your Git repository, which maybe prevented by
ingress and egress rules.