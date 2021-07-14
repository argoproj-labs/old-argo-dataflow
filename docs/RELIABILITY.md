# Reliability

Argo Dataflow has to run on Kubernetes, which means that pods can be deleted and processes killed at anytime. It avoids
using its own storage, and relies on the source or sink for storage. 

Argo Dataflow aims for **at-least once** message delivery semantics. 

The following disruptions are tolerated:

* Loss of network connection to source or sink.
* Pod deletion.
* Pipeline deletion (metrics will be lost, but no messages).

Under disruption, no messages should be lost and up to 20 messages maybe duplicated.