from argo_dataflow import pipeline, cron

if __name__ == '__main__':
    (pipeline("301-cron-log")
     .owner('argoproj-labs')
     .describe("""This example uses a cron source and a log sink.

## Cron

You can format dates using a "layout":

https://golang.org/pkg/time/#Time.Format

By default, the layout is RFC3339.

* Cron sources are **unreliable**. Messages will not be sent when a pod is not running, which can happen at any time in Kubernetes.
* Cron sources must not be scaled to zero. They will stop working.

## Log

This logs the message.

* Log sinks are totally reliable.
""")
     .step(
        (cron('*/3 * * * * *', layout='15:04:05')
         .cat('main')
         .log())
    ).save())
