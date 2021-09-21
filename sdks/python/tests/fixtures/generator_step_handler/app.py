from argo_dataflow_sdk import ProcessHandler
import time


def generator_handler():
    i = 0
    while True:
        print('running generator fn', i)
        yield f'Some Value {i}'.encode('UTF-8')
        i = i + 1
        time.sleep(1)


if __name__ == '__main__':
    processHandler = ProcessHandler()
    processHandler.start_generator(generator_handler)
