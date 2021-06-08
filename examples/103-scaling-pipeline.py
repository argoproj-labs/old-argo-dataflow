from dsls.python import pipeline, kafka

if __name__ == '__main__':
    (pipeline("103-scaling")
     .owner('argoproj-labs')
     .describe("""This is an example of having multiple replicas for a single step.

Steps can be manually scaled using `kubectl`:

```
kubectl scale step/scaling-main --replicas 3
```""")
     .step(
        (kafka('input-topic')
         .cat('main')
         .kafka('output-topic'))
    )
     .save())
