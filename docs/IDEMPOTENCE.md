# Idempotence

Because network connections can always fail to either deliver a message, or receive an acknowledgement, processors and
sinks may receive the same message twice.

It maybe that you don't mind the occasional duplicate messages. But maybe that is not the case. If so, processors and
external systems, need to be written to cater for this, to be idempotent.

The safest way to do that is for each message to contain a unique identifier that is intrinsic to the message. For
example:

* If each message is the result of a horse race, the the location and time would be enough to de-duplicate messages.
* If each message was the final year grade of a student, then the year, class, and the student's IDs would be suitable.

When there is no intrinsic identifier, then you can use the identifiers that are added as [meta-data](META.md) to each
message. You should add an identifier to you messages as soon as possible.

Some sinks have inherent idempotence, e.g. when sinking to a volume, if duplicate processing results in a file being
created with the same name, the the old file will be overwritten.