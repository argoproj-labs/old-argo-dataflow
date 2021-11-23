# Git Step

This intended as a convenient way to write steps without having to build and publish images.

When a steps starts, the code is checked out from Git, and then run:

```yaml
git:
  branch: main
  path: examples/git
  image: golang:1.17
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

## Private Repos

You'll probably need to use an SSH private key to access these. Create one as usual, e.g. by following [the Github guide](https://docs.github.com/en/github/authenticating-to-github/connecting-to-github-with-ssh/generating-a-new-ssh-key-and-adding-it-to-the-ssh-agent).

You'll then need to add this to your set-up:

```yaml
kind: Pipeline
apiVersion: dataflow.argoproj.io/v1alpha1
metadata:
  name: private-repo
spec:
  steps:
    - name: main
      git:
        url: git@github.com:alexec/private-repo.git
        image: golang:1.17
        path: .
        branch: main
        sshPrivateKeySecret:
          name: private-repo
          key: private-key
```

You'll need a secret for the private key:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: private-repo
stringData:
  private-key: |
    -----BEGIN OPENSSH PRIVATE KEY-----
    b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
    ...
    IPTRDBXMDqS0R/4BAAAAGGFjb2xsaW5zOEBpbnR1aXRkZXBlMWE5NQEC
    -----END OPENSSH PRIVATE KEY-----
```

You'll finally need to make sure that the host name ("github.com" in the above example) is in the secret name ssh:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: ssh
stringData:
  # add new hosts using `ssh-keyscan github.com`
  known_hosts: |
    # github.com:22 SSH-2.0-babeld-83b59434
    # github.com:22 SSH-2.0-babeld-83b59434
    github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==
    # github.com:22 SSH-2.0-babeld-83b59434
```

Alternatively, you can use username & password (auth token) authentication to clone a repository:

```yaml
kind: Pipeline
apiVersion: dataflow.argoproj.io/v1alpha1
metadata:
  name: private-repo
spec:
  steps:
    - name: main
      git:
        url: git@github.com:alexec/private-repo.git
        image: golang:1.17
        path: .
        branch: main
        usernameSecret:
          name: username
          key: github-access-credentials
        passwordSecret:
          name: accesstoken
          key: github-access-credentials
```

You'll need a secret for the access credentials:

```yaml
kind: Secret
apiVersion: v1
metadata:
  name: github-access-credentials
stringData:
  username: |
    <your-github-username>
  accesstoken: |
    <your-github-access-token>
```