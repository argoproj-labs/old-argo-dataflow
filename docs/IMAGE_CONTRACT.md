# Image Contract

For an image to be run as a step, it must obey the image contract.

⚠️ The contract is non-trivial to implement. Graceful handling of the SIGTERM signals is mandatory.

It must implement the following endpoints:

* http://localhost:8080/ready - must return 204 to a GET whenever it is ready to receive messages, it should not return
  204 when it is un-ready.
* http://localhost:8080/messages - must return either 204, or 201 OK to a POST (where the post body is the message
  bytes) whenever is successfully accepts a message. If it return any other code, then the message will be marked as
  errored. If it return 201, it must return the data as the HTTP response body.

It may POST a message (as bytes) to http://localhost:3569/messages and this will be sent to each sink. This endpoint
will return standard HTTP response codes, including 500 if the message could not be processed.

The container will be started with an file `/var/run/argo-dataflow/authorization`. The string value is this must be
passed to `/messages` as a `Authorization: $(cat /var/run/argo-dataflow/authorization)`.

It must gracefully shutdown when SIGTERM on PID 1 is executed in the container, specifically respond to in-flight
requests and become un-ready.

⚠️ This is not quite the same as a SIGTERM it will get from the Kubelet on pod deletion. The image must obey that too.

## Unix Domain Socket (UDS)

UDS are about 30% faster that TCP sockets. An image may optionally create a UDS at `/var/run/argo-dataflow/main.sock`
rather listening on port 8080. 