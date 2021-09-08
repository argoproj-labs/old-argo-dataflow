from asyncio.exceptions import CancelledError
from argo_dataflow_sdk import ProcessHandler
import asyncio

async def generator_handler():
  try:
    i = 0
    while True:
      print('running generator fn', i)
      yield f'Some Value {i}'.encode('UTF-8')
      i = i + 1
      await asyncio.sleep(1)
  except CancelledError:
    print('Generator function got cancelled, time to cleanup.')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start_generator(generator_handler)
