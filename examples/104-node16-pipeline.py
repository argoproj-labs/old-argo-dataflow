from argo_dataflow import pipeline, kafka


def handler(msg, context):
    return ("hi! " + msg.decode("UTF-8")).encode("UTF-8")


if __name__ == '__main__':
    (pipeline("104-node16")
     .owner('argoproj-labs')
     .describe("""This example is of the NodeJS 16 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .step(
        (kafka('input-topic')
         .code(code="""module.exports = async function (messageBuf, context) {
  const msg = messageBuf.toString('utf8')
  return Buffer.from('hi ' + msg)
}""", runtime='node16')
         .kafka('output-topic')
         ))
     .save())
