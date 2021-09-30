def handler(message, context):
    msg = message.decode("UTF-8")
    print('Got message', msg)
    return ("hi " + msg).encode('UTF-8')
