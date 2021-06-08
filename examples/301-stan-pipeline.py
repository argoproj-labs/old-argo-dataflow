from dsls.python import pipeline, stan

if __name__ == '__main__':
    (pipeline("301-stan")
     .describe("""This example shows reading and writing to a STAN subject""")
     .step(
        (stan('input-subject')
         .cat('main')
         .stan('output-subject')
         ))
     .save())
