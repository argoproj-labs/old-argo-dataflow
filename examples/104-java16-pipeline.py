from dsls.python import pipeline, kafka


def handler(msg):
    return msg


if __name__ == '__main__':
    (pipeline("104-java-16")
     .describe("""This example is of the Java 16 handler.

[Learn about handlers](../docs/HANDLERS.md)""")
     .step(
        (kafka('input-topic')
         .handler('main', code="""public class Handler {
    public static byte[] Handle(byte[] msg) throws Exception {
        return msg;
    }
}""", runtime='java16')
         .kafka('output-topic')
         ))
     .save())
