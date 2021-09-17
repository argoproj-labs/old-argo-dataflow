from argo_dataflow_sdk import ProcessHandler
import asyncio

async def generator_handler():
    i = 0
    while True:
      print('running generator fn', i)
      yield f'Some Value {i}'.encode('UTF-8')
      await asyncio.sleep(0.1)
      raise ValueError('Some error')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start_generator(generator_handler)
