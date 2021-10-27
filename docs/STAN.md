# STAN

If you want to experiment with STAN, install STAN:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/config/apps/stan.yaml
```

Configure dataflow to use that STAN by default:

```bash
kubectl apply -f https://raw.githubusercontent.com/argoproj-labs/argo-dataflow/main/examples/dataflow-stan-default-secret.yaml 
```

Wait for the statefulsets to be available (ctrl+c when available):

```bash
kubectl get statefulset -w
```
