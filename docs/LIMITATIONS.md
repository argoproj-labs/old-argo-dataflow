# Limitations

## Message Size

Messages are handled in memory by Argo Dataflow, so are limited by the amount of memory available x the number of messages cached in memory. 

* HTTP messages must be < 4GB.
* Kafka messages are typically < 1MB. 
* NATS streaming messages are < 1MB.

## Message Throughput

* HTTP source tested to 2,000 TPS
* Sinking is limited by the rate of the sink.