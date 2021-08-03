from argo_dataflow import pipeline, cron, http

if __name__ == '__main__':
    (pipeline("301-http")
     .owner('argoproj-labs')
     .describe("""This example uses a HTTP sources and sinks.

HTTP has the advantage that it is stateless and therefore cheap. You not need to set-up any storage for your
messages between steps. 

* A HTTP sink may return and error code, clients will need to re-send messages.
* Adding replicas will nearly linearly increase throughput.
* Due to the lack of state, do not use HTTP sources and sinks to connect steps. 
""")
     .annotate("dataflow.argoproj.io/needs", "dataflow-103-http-main-source-default-secret.yaml")
     .step(
        (cron('*/3 * * * * *')
         .cat('cron')
         .http('http://http-main/sources/default')
         ))
     .step(
        (http(serviceName='http-main')
         .cat('main')
         .log()
         ))
     .save())
