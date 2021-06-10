# Reliability

Argo Dataflow has to run on Kubernetes, which means that pods can be deleted and processes killed at anytime. It avoids
using its own storage, and relies on the source or sink for storage. Data can be lost in the following ways:

* Failure to ack source messages as processed.
* Failure of main container to process messages.
* Failure to sink message.

These can all because be issues such as crashed on network issues.

Messages are only acked once they have been successfully sunk.

This means that Argo Dataflow has at-least once message delivery semantics. 