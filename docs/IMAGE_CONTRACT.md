# Image Contract

For an image to be run as a step, it must obey the image contract.

It must implement the following endpoints:

* http://localhost:8080/ready - must return 204 to a GET whenever it is ready to receive messages, it should not return 204 when it is not ready.
* http://localhost:8080/messages - must return either 204, or 201 OK to a POST (where the post body is the message bytes) whenever is successfully accepts a message. If it return any other code, then the message will be marked as errored. If it return 201, it must return the data as the HTTP response body. 

It may POST a message (as bytes) to http://localhost:3569/messages and this will be sent to each sink. This endpoint will return standard HTTP response codes, including 500 if the message could not be processed. 