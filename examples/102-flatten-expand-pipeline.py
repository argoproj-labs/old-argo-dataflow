from dsls.python import cron, pipeline, stan

if __name__ == '__main__':
    (pipeline("102-flatten-expand")
     .owner('argoproj-labs')
     .describe("""This is an example of built-in flattening and expanding.""")
     .step(
        cron('*/3 * * * * *')
            .map('generate', """bytes('{"foo": {"bar": "' + string(msg) + '"}}')""")
            .stan('data'))
     .step(
        stan('data')
            .flatten('flatten')
            .stan('flattened')
    )
     .step(
        stan('flattened')
            .expand('expand')
            .log()
    )
     .save())
