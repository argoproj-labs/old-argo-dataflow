from argo_dataflow import pipeline, kafka

if __name__ == '__main__':
    (pipeline("102-map")
     .owner('argoproj-labs')
     .describe("""This is an example of built-in mapping.

Maps are written using expression syntax and must return a byte array.

They have a single variable, `msg`, which is a byte array.

[Learn about expressions](../docs/EXPRESSIONS.md)""")
     .step(
        (kafka('input-topic')
         .map(expression="bytes('hi ' + string(msg))")
         .kafka('output-topic'))
    )
        .save())
