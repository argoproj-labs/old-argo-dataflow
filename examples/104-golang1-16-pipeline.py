from argo_dataflow import pipeline, kafka


def handler(msg):
    return msg


if __name__ == '__main__':
    (pipeline("104-golang1-16")
     .owner('argoproj-labs')
     .describe("""This example of Go 1.16 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .step(
        (kafka('input-topic')
         .code('main', code="""package main

import "context"

func Handler(ctx context.Context, m []byte) ([]byte, error) {
  return []byte("hi " + string(m)), nil
}""", runtime='golang1-16')
         .log()
         ))
     .save())
