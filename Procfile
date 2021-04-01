create-input-topic: go run ./kafka -topic input-topic create-topic
create-output-topic: go run ./kafka -topic output-topic create-topic
runner: make runner
kafka-9092: make kafka-9092
nats-4222: make nats-4222
nats-8222: make nats-8222
controller: go run ./main.go -metrics-addr :7070
watch: kubectl get pl,step -w
logs: make logs
input: go run ./kafka -topic input-topic pump-topic
output: kafka-console-consumer -topic output-topic
