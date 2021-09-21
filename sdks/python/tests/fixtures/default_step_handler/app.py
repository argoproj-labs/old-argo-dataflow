from argo_dataflow_sdk import ProcessHandler


def handler(message, _):
    msg = message.decode("UTF-8")
    return ("Hi " + msg).encode('UTF-8')


if __name__ == '__main__':
    processHandler = ProcessHandler()
    processHandler.start(handler)
