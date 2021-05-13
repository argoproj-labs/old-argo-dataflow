runner: make runner
runtimes: make runtimes
controller: ARGO_DATAFLOW_PULL_POLICY=IfNotPresent ARGO_DATAFLOW_INSTALLER=true ARGO_DATAFLOW_UPDATE_INTERVAL=5s ARGO_DATAFLOW_NAMESPACE=argo-dataflow-system go run ./manager -metrics-addr :7070
logs: make logs
argocli: make argocli
ui: make ui
