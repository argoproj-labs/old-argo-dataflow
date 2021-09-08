# Argo_Dataflow_Sdk

## Project setup

```bash
make env
make install
```

## Running tests

```bash
make test
```

## Using the SDK

### Executing an Argo-Dataflow Step with a source and a sink

```python
from argo_dataflow_sdk import ProcessHandler

def handler(message, _):
  msg = message.decode("UTF-8")
  return ("Hi " + msg).encode('UTF-8')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
```

Or as async python function

```python
from argo_dataflow_sdk import ProcessHandler
import asyncio

async def handler(message, _):
  msg = message.decode("UTF-8")
  await asyncio.sleep(1)
  return ("Hi " + msg).encode('UTF-8')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
```

### Executing an Argo-Dataflow Step with only source

```python
from argo_dataflow_sdk import ProcessHandler

def handler(message, _):
  msg = message.decode("UTF-8")
  print('Got message', msg)

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
```

### Executing an Argo-Dataflow Generator Step with only sink

```python
from asyncio.exceptions import CancelledError
from argo_dataflow_sdk import ProcessHandler
import time

def send_message_every_second():
    i = 0
    while True:
      print('running generator fn', i)
      yield f'Some Value {i}'.encode('UTF-8')
      i = i + 1
      time.sleep(1)

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start_generator(send_message_every_second)
```

Or as async generator handler

```python
from asyncio.exceptions import CancelledError
from argo_dataflow_sdk import ProcessHandler
import asyncio

async def send_message_every_second():
    i = 0
    while True:
      print('running generator fn', i)
      yield f'Some Value {i}'.encode('UTF-8')
      i = i + 1
      await asyncio.sleep(1)

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start_generator(send_message_every_second)
```

### Error handling in Argo-Dataflow Step

An error like this one:

```python
from argo_dataflow_sdk import ProcessHandler

def handler(message, _):
  raise ValueError('Some error')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
```

Will cause the message to be marked as failed processing, but the step will keep running, and will be able to process new messages.

In asyncio version of this step, error handling works the same way.

### Error handling in Argo-Dataflow Generator Step

An error like this one:

```python
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
```

Is considered a critical step failure, and will cause `processHandler.start_generator(generator_handler)` to return a None. In most cases this will mean shutting down the whole container that is running this step. That way, failure should be visible in Argo-Dataflow UI.

In asyncio version of this step, error handling works the same way.
