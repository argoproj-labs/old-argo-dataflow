import asyncio


async def handler():
    i = 0
    while True:
        print('running generator fn', i)
        yield f'Some Value {i}'.encode('UTF-8')
        i = i + 1
        await asyncio.sleep(1)
