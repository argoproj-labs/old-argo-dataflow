create-input-topic: go run ./topic-creator -topic input-topic
create-output-topic: go run ./topic-creator -topic output-topic
runner: make runner
kafka-9092: make kafka-9092
nats-4222: make nats-4222
nats-8222: make nats-8222
controller: go run ./main.go -metrics-addr :7070
watch: kubectl get pl,fn -w
logs: make logs
input: seq 9999999 | while read i ; do kafka-console-producer -topic input-topic -value my-val-$i ; done
output: kafka-console-consumer -topic output-topic
