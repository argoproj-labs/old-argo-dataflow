from dsls.python import pipeline, cron, http

if __name__ == "__main__":
    (pipeline("http")
     .describe("""This example uses a HTTP sources and sinks.

HTTP has the advantage that it is stateless and therefore cheap. You not need to set-up any storage for your
messages between steps. Unfortunately, it is possible for some or all of your messages to not get delivered.
Also, this is sync, not async, so it can be slow due to the time taken to deliver messages.""")
     .annotate("dataflow.argoproj.io/test", "true")
     .step(
        (cron('*/3 * * * * *', layout= "15:04:05")
         .cat('cron')
         .http('http://http-main/sources/default')
         ))
     .step(
        (http()
         .cat('main')
         .log()
         ))
     .dump())
