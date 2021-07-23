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