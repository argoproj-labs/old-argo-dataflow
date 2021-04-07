# Proposal

## Collaborators and Consulted

To discus:

* Intuit
* RedHat
* Argo Community
* CNCF Serverless Workflows
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
* Running reliable processing on unreliable infrastructure. Pods can be killed at anytime.
* Scale-to-zero, queue metrics based scaling.
* Data storage.

Argo's areas of strength:

* Is cloud-native
* Powerful user interface
* Easy to get started quickly
* Strong community
* Ecosystem effects

The unit of processing in Kubernetes is a container. The unit of scaling is a pod.

So:

* `Pipeline` maps to a custom resource.
* `ParDo` maps to one or more running container images that can be scaled by running more containers in the pod, and by
  running more pods.
* `GroupByKey` must be done by the infrastructure. Each container get a window of keys.
* `PCollection` TBD
* `Window` We need to figure out how to window data.

### Data Format

Similar to CloudEvents. Enable easy interop with other compliant tools.

### Features

* Each step is a container - deploy an image, or have DF build your code from Git
* Connect steps to send and receive images from Kafka and NATS

### Data Input/Output Options

#### HTTP

Slower, but easier to get right.

To receive messages from sinks via HTTP, expose an endpoint on `http://localhost:8080/messages`. Dataflow will POST each
message to it.

To send messages to source via HTTP, POST the message to `http://localhost:3569/messages`.

#### FIFO

Core Linux capability for IPC.

Read messages from `/var/run/argo-dataflow/in` and write them to `/var/run/argo-dataflow/out`. Each messages must be a
single line - you must escape new lines.

#### Future

* files - chunky - but great for grouping by key
* socket - fast, but the low level programming is hard to get right
* stdin/stdout - performance can be poor on these - can be achived more easily and flexible using FIFO

We may well want several options.

### Data Chunking (Future)

After discussing with @vigith, we need to support this. We need to provide a way for a user to decide when a chunk
starts and end, this maybe a function itself. A pod will receive a sequence of messages as a series of frames, e.g.

1. `type=ChunkStart`
1. `type=Data, payload=...`
1. `type=Data, payload=...`
1. ...
1. `type=ChunkEnd`

Messages could also be chunk-less and the `Chunk*` control messages are never sent.

