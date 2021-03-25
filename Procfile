create-input-topic: go run ./topic-creator -topic input-topic
create-output-topic: go run ./topic-creator -topic output-topic
sidecar-image: make sidecar-image
cat-image: make cat-image
kafka: make kafka
input: seq 9999999 | while read i ; do kafka-console-producer -topic input-topic -value my-val-$i ; sleep 1; done
output: kafka-console-consumer -topic output-topic
controller: go run ./main.go
logs: make logs
watch-pipelines: kubectl get pipeline -w