# Meta-data

Messages come with meta-data that can be accessed via the messages context:

| Name | Description |
|---|---|
| `source` | A URN for the source the message came from |
| `id` | A unique identifier for the messages within the source |

`source+id` is intended to be globally unique.

Source:

* Cron: `urn:dataflow:cron:${schedule}`
* Database: `urn:dataflow:db:${dbURL}` or `urn:dataflow:db:${secret}.secret.${namespace}.${cluster}`
* HTTP: `urn:dataflow:http:https://${serviceName}.svc.${namespace}.${cluster}` or `urn:dataflow:http:${endpoint}`
* Kafka: `urn:dataflow:kafka:${broker[0]}:${topic}`
* S3: `urn:dataflow:s3:${bucket}`
* STAN: `urn:dataflow:stan:${natsURL}:${subject}`
* NATS JetStream `urn:dataflow:jetstream:${natsURL}:${subject}`
* Volume:
    * `urn:dataflow:volume:configmap:${configmap}.configmap.${namespace}.${cluster}`
    * `urn:dataflow:volume:secret:${secret}.secret.${namespace}.${cluster}`

IDs:

* Cron: `${now}`
* Database: `${offset}`
* HTTP: `${randomGUID}`
* Kafka: `${partition}-${offset}`
* S3: `${key}`
* STAN: `${sequence}`
* NATS JetStream: `${consumer.sequence}-${stream.sequence}`
* Volume: `${filename}`
