from argo_dataflow_sdk import ProcessHandler

def generator_handler():
    i = 0
    while True:
      print('running generator fn', i)
      yield f'Some Value {i}'.encode('UTF-8')

      raise ValueError('Some error')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start_generator(generator_handler)
