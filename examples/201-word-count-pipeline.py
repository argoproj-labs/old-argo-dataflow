from argo_dataflow import pipeline, kafka, stan

if __name__ == '__main__':
    (pipeline("201-word-count")
     .owner('argoproj-labs')
     .describe("""This pipeline count the number of words in a document, not the number of count of each word as you might expect.

  It also shows an example of a pipelines terminates based on a single step's status.""")
     .annotate("dataflow.argoproj.io/test", "false")
     .annotate("dataflow.argoproj.io/needs", "word-count-input-configmap.yaml")
     .step(
        (kafka('input-topic')
         .container('read-text',
                    args=[
                        'sh', '-c', """cat /in/text | tee -a /var/run/argo-dataflow/out"""],
                    image='ubuntu:latest',
                    volumeMounts=[{'name': 'in', 'mountPath': '/in'}],
                    volumes=[{'name': 'in', 'configMap': {
                        'name': 'word-count-input'}}]
                    )
         .stan('lines')
         ))
     .step(
        (stan('lines')
         .container('split-lines',
                    args=['bash', '-c', """set -eux -o pipefail
while read sentence; do
  for word in $sentence; do
      echo $word
  done
done > /var/run/argo-dataflow/out < /var/run/argo-dataflow/in"""],
                    image='ubuntu:latest',
                    fifo=True,
                    )
         .scale('minmax(pending, 0, 1)')
         .stan('words')
         ))
     .step(
        (stan('words')
         .container('count-words',
                    args=['bash', '-c', """set -eux -o pipefail
i=0
while read word && [ $word != EOF ]; do
  i=$((i+1))
done < /var/run/argo-dataflow/in
echo $i > /var/run/argo-dataflow/out"""],
                    image='ubuntu:latest',
                    fifo=True,
                    )
         .terminator()
         .kafka('output-topic')
         ))
     .save())
