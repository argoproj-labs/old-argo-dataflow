from dsls.python import pipeline, kafka


def handler(msg):
    return msg


if __name__ == '__main__':
    (pipeline("go1-16")
     .describe("""This example of Go 1.16 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .step(
        (kafka('input-topic')
         .handler('main', code="""package main

func Handler(m []byte) ([]byte, error) {
  return []byte("hi " + string(m)), nil
}""", runtime='go1-16')
         .log()
         ))
     .dump())
