from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("103-autoscaling")
     .owner('argoproj-labs')
     .describe("""This is an example of having multiple replicas for a single step.

Replicas are automatically scaled up and down depending on the the desired formula, which can be computed using the following:

* `P` total number of pending messages.
* `p` change in number of pending messages.
* `c` the current number of replicas.
* `limit(v, min, max, delta)` a function to constraint the minimum and maximum number of replicas, as well as the step-up/down.

In this example:

* Each period is 60s.
* Each replica can consume 250 messages each second.
* We want to consume all pending messages in 10 periods.
* We want to have between 0 and 4 replicas, and scale-up or down maximum 2 replicas at a time. 

### Scale-To-Zero and Peeking

You can scale to zero. The number of replicas will be periodically scaled
to 1 so it can "peek" the the message queue. The number of pending messages is measured and the target number
of replicas re-calculated.""")
     .step(
        (kafka('input-topic')
         .cat('main')
         .scale('limit(c + p / (60 * 250) + P / (10 * 60 * 250), 0, 4, 2)', scalingDelay='"1m"', peekDelay='"20m"')
         .kafka('output-topic'))
    )
     .save())
