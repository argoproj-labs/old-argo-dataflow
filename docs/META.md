# Meta-data

Messages come with meta-data that can be accessed via the messages context:

| Name | Description |
|---|---|
| `source` | A URN for the source the message came from |
| `id` | A unique identifier for the messages within the source |

`source+id` is intended to be globally unique.

Source:

* Cron: `urn:dataflow:cron:${schedule}`
* Database: `urn:dataflow:db:${dbURL}` or `urn:dataflow:db:${secret}.pod.${namespace}.${cluster}`
* HTTP: `urn:dataflow:http:https://${serviceName}.svc.${namespace}.${cluster}` or `urn:dataflow:http:${endpoint}`
* Kafka: `urn:dataflow:kafka:${broker[0]}:${topic}`
* S3: `urn:dataflow:s3:${bucket}`
* STAN: `urn:dataflow:stan:${natsURL}:${subject}`
* Volume:
    * `urn:dataflow:volume:emptydir:${pod}.pod.${namespace}.${cluster}`
    * `urn:dataflow:volume:secret:${secret}.secret.${namespace}.${cluster}`

IDs:

* Cron: `${now}`
* Database: `${offset}`
* HTTP: `$randomGUID}`
* Kafka: `${partiton}-${offset}`
* S3: `${key}`
* STAN: `${sequence}`
* Volume: `${filename}`
