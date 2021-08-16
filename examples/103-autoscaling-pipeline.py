from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("103-autoscaling")
     .owner('argoproj-labs')
     .describe("""This is an example of having multiple replicas for a single step.

Replicas are automatically scaled up and down depending on the the desired formula, which can be computed using the following:

* `P` total number of pending messages.
* `p` change in number of pending messages
* `m` change in total number of of consumed messages.
* `c` the current number of replicas.
* `minmax(v, min, max)` a function to constraint the minimum and maximum number of replicas.

### Scale-To-Zero and Peeking

You can scale to zero by setting `minReplicas: 0`. The number of replicas will start at zero, and periodically be scaled
to 1  so it can "peek" the the message queue. The number of pending messages is measured and the target number
of replicas re-calculated.""")
     .step(
        (kafka('input-topic')
         .cat('main')
         .scale('minmax((m + p) * c / m + P / (c * 10 * m), 0, 4)')
         .kafka('output-topic'))
    )
     .save())
