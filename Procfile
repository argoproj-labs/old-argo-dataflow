kafka: make kafka
create-input-topic: go run ./topic-creator -topic input-topic
create-output-topic: go run ./topic-creator -topic output-topic
cat-image: make cat-image
init-image: make init-image
sidecar-image: make sidecar-image
controller: go run ./main.go -metrics-addr :7070
logs: make logs
nats: make nats
watch-pipelines: kubectl get pipeline -w
input: seq 9999999 | while read i ; do kafka-console-producer -topic input-topic -value my-val-$i ; done
output: kafka-console-consumer -topic output-topic
