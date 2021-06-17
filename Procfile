controller: ARGO_DATAFLOW_PULL_POLICY=IfNotPresent ARGO_DATAFLOW_UPDATE_INTERVAL=10s ARGO_DATAFLOW_NAMESPACE=argo-dataflow-system go run ./manager
logs: make logs
argocli: make argocli
ui: make ui
test: kubectl port-forward svc/testapi 8378:8378