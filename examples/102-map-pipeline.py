from dsls.python import pipeline, kafka

if __name__ == '__main__':
    (pipeline("map")
     .describe("""This is an example of built-in mapping.

Maps are written using expression syntax and must return a byte array.

They have a single variable, `msg`, which is a byte array.

[Learn about expressions](../docs/EXPRESSIONS.md)""")
     .step(
        (kafka('input-topic')
         .map('main', "bytes('hi ' + string(msg))")
         .kafka('output-topic'))
    )
     .dump())
