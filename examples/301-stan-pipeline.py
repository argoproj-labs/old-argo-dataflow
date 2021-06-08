from dsls.python import pipeline, kafka, stan

if __name__ == '__main__':
    (pipeline("stan")
     .describe("""This example shows reading and writing to a STAN subject""")
     .step(
        (stan('input-subject')
         .cat('main')
         .stan('output-subject')
         ))
     .dump())
