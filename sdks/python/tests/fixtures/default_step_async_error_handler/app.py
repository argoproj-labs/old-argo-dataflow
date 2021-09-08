from argo_dataflow_sdk import ProcessHandler
from aiohttp import ClientSession

async def handler(message, _):
  msg = message.decode("UTF-8")
  async with ClientSession() as session:
    async with session.get('http://localhost:8080/ready') as response:
      body = (await response.content.read()).decode('utf-8')
      assert response.status == 203
      assert body == ''
  return ("Hi " + msg).encode('UTF-8')

if __name__ == '__main__':
  processHandler = ProcessHandler()
  processHandler.start(handler)
