controller: source manager.env && go run -race -ldflags="-X 'github.com/argoproj-labs/argo-dataflow/shared/util.version=v0.0.0-latest-0'" ./manager
logs: make logs
argocli: [ "$UI" = true ] && make argocli
ui: [ "$UI" = true ] && make ui