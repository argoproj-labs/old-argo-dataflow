from dsls.python import pipeline, kafka

if __name__ == "__main__":
    (pipeline("group")
     .describe("""This is an example of built-in grouping.

WARNING: The spec/syntax not been finalized yet. Please tell us how you think it should work!

There are four mandatory fields:

* `key` A string expression that returns the message's key
* `endOfGroup` A boolean expression that returns whether or not to end group and send the group messages onwards.
* `format` What format the grouped messages should be in.
* `storage` Where to store messages ready to be forwarded.

[Learn about expressions](../docs/EXPRESSIONS.md)

### Storage

Storage can either be:

* An ephemeral volume - you don't mind loosing some or all messages (e.g. development or pre-production).
* A persistent volume - you want to be to recover (e.g. production).""")
     .step(
        (kafka('input-topic')
         .group('main',
                key='string(msg) contains "2" ? "even" : "odd"',
                format='JSONStringArray',
                endOfGroup='string(msg) contains "excited"',
                storage={'emptyDir': {}}
                )
         .stan('odd-end')
         ))
     .dump())
