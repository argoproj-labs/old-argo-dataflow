# Scaling

You can scale in the following ways:

* Using the built-in scaling, as shown in 103-replicas-pipeline.yaml
* Using `kubect scale step/{pipelineName}-{stepName}` --replicas 1
* Using a [Horizontal Pod Autoscaler](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/).
