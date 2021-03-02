# Argo Streams

## Primer Reading

* [Streaming 101: The world beyond batch](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-101) O'Reilly blog post
* [Streaming 102: The world beyond batch](https://www.oreilly.com/radar/the-world-beyond-batch-streaming-102) O'Reilly blog post
* [The Dataflow Model: A Practical Approach to Balancing Correctness, Latency, and Cost in Massive-Scale, Unbounded, Out-of-Order Data Processing](http://www.vldb.org/pvldb/vol8/p1792-Akidau.pdf) Google whitepaper(?)

## Use Cases

* Real-time "click" analytics
* Anomoly detection
* Fraud detection
* Operational (including IoT) analytics

## Cloud Provider Solutions

*  [Google Cloud Dataflow](https://cloud.google.com/dataflow)
*  [Amazon Kenisis](https://aws.amazon.com/kinesis/) - including Data Streams, Data Firehose, and Data Analytics
*  [Azure Stream Analytics](https://azure.microsoft.com/en-us/services/stream-analytics/)

## Collabarators and Consulted

To discus:

* Intuit
* RedHat
* Argo Community
* Google TektonCD team
* Kubeflow Pipelines team

## Proposal

* Cloud-native
* Language-agnostic

Relevant Kubernetes benefits:

* Built-in scaling support, e.g. HPAs, VPAs, or plain deployments.
* Declarative and GitOps friendly.

Kubernetes challenges:

* Friction when running task-based workloads on an application based platform.
* Running reliabile processing on unreliable infrastructure. Pods can be killed at anytime.
* Scale-to-zero, queue metrics based scaling.
* Data storage.

Argo areas of strength:

* Is cloud-native
* Powerful user interface
* Easy to get started quickly
* Strong community
* Ecosystem effects

The unit of procssing in Kubernetes is a container. The unit of scaling is a pod.

So:

* `Pipeline` maps to a custom resource.
* `ParDo` maps to one or more running container images that can be scaled by runing more containers in the pod, and by running more pods.
* `GroupByKey` must be done by the infrastructure. I.e. each container would expect to get specific set of keys.
* `PCollection` TBD
* `Window` Do we want to support this in v1? It is a hard problem.


```yaml
apiVersion: argoproj.io/v1alpha1
kind: Pipeline
metadata:
  name: work-time
spec:
  - kafka:
      topic: taxirides
  - groupByKey:
      key: .licenseId
  - parDo: some-combiner-image
  - parDo: some-logger-image
```

`ParDo` has an input stream of data and outputs another. What does this look like in container world? 

* stdin/stdout - performance can be poor on these
* named pipes - not commonly used, but core Linux capability for IPC
* files - not very "streamy"
* socket - fast, but the low level programming is hard to get right
* HTTP endpoints - slower, but easier to get right

## Further Reading

* [Beam Capability Matrix
](https://beam.apache.org/documentation/runners/capability-matrix/)
