# Cluster Quick Start
 The Cluster scope quick start will deploy  the Argo Dataflow controller in cluster scope to execute the pipeline on 
 all namespace in that cluster.
 
 You'll need a Kubernetes cluster to try it out, e.g. Docker for Desktop.

Deploy into the `argo-dataflow-system` namespace:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/cluster-quick-start.yaml
```


Change to the installation namespace:

```bash
kubectl config set-context --current --namespace=argo-dataflow-system
```

Wait for the deployments to be available (ctrl+c when available):

```bash
kubectl get deploy -w
```

If you want the user interface:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/apps/argo-server.yaml
kubectl get deploy -w ;# (ctrl+c when available)
kubectl port-forward svc/argo-server 2746:2746
```

To Run pipeline on other namespace, it should have `RBAC` to execute the `pipeline` on each namespace.

you have to create below  `serviceaccount`, `Role`, and `Rolebinding`. [yamls](https://github.com/argoproj-labs/argo-dataflow/tree/main/config/default-cluster)


```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pipeline
  namespace: default  # Update your namespace
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: pipeline
  namespace: default  # Update your namespace
rules:
  - apiGroups:
      - dataflow.argoproj.io
    resources:
      - steps
    verbs:
      - create
      - delete
      - get
      - list
      - update
      - watch
  - apiGroups:
      - dataflow.argoproj.io
    resources:
      - pipelines/status
    verbs:
      - update
  - apiGroups:
      - dataflow.argoproj.io
    resources:
      - steps/status
    verbs:
      - update
  - apiGroups:
      - dataflow.argoproj.io
    resources:
      - steps/scale
    verbs:
      - patch
  - apiGroups:
      - ""
    resources:
      - pods
      - pods/exec
    verbs:
      - create
      - get
      - list
      - watch
      - delete
  - apiGroups:
      - ""
    resources:
      - events
    verbs:
      - create
      - patch
  - apiGroups:
      - ""
    resources:
      - secrets
    verbs:
      - create
      - get
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pipeline
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: pipeline
subjects:
  - kind: ServiceAccount
    name: pipeline
    namespace: default # Update your namespace
```


Open [http://localhost:2746/pipelines/argo-dataflow-system](http://localhost:2746/pipelines/argo-dataflow-system).

Run "hello" from the [one of the examples](EXAMPLES.md).

Finally, you can set-up [Kafka](KAKFA.md) or [NATS Streaming](STAN.md) if you want to experiment with them as sources
and sinks.