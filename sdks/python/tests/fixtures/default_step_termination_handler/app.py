from argo_dataflow_sdk import ProcessHandler
import time

def handler(message, _):
  msg = message.decode("UTF-8")
  time.sleep(0.2)
  return ("Hi " + msg).encode('UTF-8')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
