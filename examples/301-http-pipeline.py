from dsls.python import pipeline, cron, http

if __name__ == '__main__':
    (pipeline("301-http")
     .owner('argoproj-labs')
     .describe("""This example uses a HTTP sources and sinks.

HTTP has the advantage that it is stateless and therefore cheap. You not need to set-up any storage for your
messages between steps. 

* HTTP sinks are *unreliable* because it is possible for messages to not get delivered when the receiving service is down.
""")
     .annotate("dataflow.argoproj.io/test", "true")
     .step(
        (cron('*/3 * * * * *')
         .cat('cron')
         .http('http://http-main/sources/default')
         ))
     .step(
        (http()
         .cat('main')
         .log()
         ))
     .save())
