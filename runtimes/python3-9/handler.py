def handler(msg, context):
    return ("hi! " + msg.decode("UTF-8")).encode("UTF-8")
