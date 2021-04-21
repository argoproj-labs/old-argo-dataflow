runner: make runner
runtimes: make runtimes
controller: PULL_POLICY=IfNotPresent INSTALLER=true ARGO_DATAFLOW_UPDATE_INTERVAL=5s ARGO_DATAFLOW_NAMESPACE=argo-dataflow-system go run ./main.go -metrics-addr :7070
logs: make logs
argocli: make argocli
ui: make ui
