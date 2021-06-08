from dsls.python import pipeline, kafka, stan

if __name__ == '__main__':
    (pipeline("vet")
     .describe("""This pipeline processes pets (cats and dogs).""")
     .annotate("dataflow.argoproj.io/test", "false")
     .annotate("dataflow.argoproj.io/needs", "pets-configmap.yaml")
     .step(
        (kafka('input-topic')
         .container('read-pets',
                    args=['sh', '-c', """cat /in/text | tee -a /var/run/argo-dataflow/out"""],
                    image='ubuntu:latest',
                    volumes=[{'name': 'in', 'configMap': {'name': 'pets'}}]
                    )
         .stan('pets')
         ))
     .step(
        (stan('pets')
         .filter('filter-cats', 'string(msg) contains "cat"')
         .stan('cats'))
    )
     .step(
        (stan('cats')
         .map("process-cats", 'json("Meow! " + object(msg).name)'))
            .kafka('output-topic')
    )
     .step(
        (stan('pets')
         .filter('filter-dogs', 'string(msg) contains "dog"')
         .stan('dogs'))
    )
     .step(
        (stan('dogs')
         .map("process-dogs", 'json("Woof! " + object(msg).name)'))
            .kafka('output-topic')
    )
     .dump())
