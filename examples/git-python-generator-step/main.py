from argo_dataflow_sdk import ProcessHandler

from .handler import handler

if __name__ == '__main__':
    processHandler = ProcessHandler()
    processHandler.start_generator(handler)
