from dsls.python import pipeline, kafka

if __name__ == '__main__':
    (pipeline("108-fifos")
     .owner('argoproj-labs')
     .describe("""This example use named pipe to send and receive messages.

Two named pipes are made available:

* The container can read lines from `/var/run/argo-dataflow/in`. Each line will be a single message.
* The contain can write to `/var/run/argo-dataflow/out`. Each line MUST be a single message.

You MUST escape new lines.""")
     .step(
        (kafka('input-topic')
         .container('main',
                    args=['sh', '-c', """cat /var/run/argo-dataflow/in | while read line ; do
  echo "hi $line"
done > /var/run/argo-dataflow/out"""],
                    image='ubuntu:latest',
                    fifo=True
                    )
         .kafka('output-topic')
         ))
     .save())
