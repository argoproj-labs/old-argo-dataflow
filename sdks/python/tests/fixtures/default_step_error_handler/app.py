from argo_dataflow_sdk import ProcessHandler


def handler(message, _):
    raise ValueError('Some error')


if __name__ == '__main__':
    processHandler = ProcessHandler()
    processHandler.start(handler)
