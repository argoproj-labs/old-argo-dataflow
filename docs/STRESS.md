# Stress

Create an appropriately sized cluster.

Turn off Netskope if you're running on GCP. 

Run `caffeinate` in a new terminal to prevent your Mac going to sleep.

Check you can connect with `kubectl cluster-info`.

Install using `make deploy`.

Set namespace `kubens argo-dataflow-system`.

Scale-up the controller `kubectl scale deploy/controller-manager --replicas 1`.

Check everything is ready: `make wait`.

Look at test/stress/params.go to choose your parameters you have.

Run the test.

Check test/stress/test-results.json for the results.

# Large TPS Stress Testing

Upto 1M TPS tested on http/kafka/stan sources on v0.0.106. 

### HTTP:

Run ` env MESSAGE_SIZE=100 N=1000000 TIMEOUT=30m WORKERS=299  make test-http-stress`

### Kafka:
Run ` env MESSAGE_SIZE=100 N=1000000 TIMEOUT=30m make test-kafka-stress`    

### Stan:
Run ` env MESSAGE_SIZE=100 N=1000000 TIMEOUT=30m make test-stan-stress`    

Results:

|Source|	50K TPS |	100K TPS|	500K TPS	|1M TPS|
|---|---|---|---|---|  
|HTTP	|TPS 5500	|TPS 17232	|TPS 33864	|TPS 65384|
|Kafka	|TPS 2040	|TPS 1993	|TPS 2457	|TPS 1755|
|Stan	|TPS 1284	|TPS 1356|		