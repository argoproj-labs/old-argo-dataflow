from asyncio import iscoroutinefunction, create_task
import asyncio
import sys
from os import getpid, environ
from types import AsyncGeneratorType, GeneratorType
from os.path import exists, isfile
import logging

from aiohttp import web, ClientSession

if environ.get('PYTHONDEBUG'):
    logging.basicConfig(level=logging.DEBUG)

HOST_NAME = "0.0.0.0"
SERVER_PORT = 8080
GENERATOR_STEP_SINK = 'http://localhost:3569/messages'
AUTH_FILE_PATH = environ.get(
    'AUTH_FILE', '/var/run/argo-dataflow/authorization')


class ProcessHandler:
    handler = None

    async def __ready_handler(self, _):
        return web.Response(status=204)

    async def __messages_handler(self, request):
        try:
            msg = await request.content.read()
            out = None

            if iscoroutinefunction(self.handler):
                out = await self.handler(msg, {})
            else:
                out = self.handler(msg, {})

            if out:
                return web.Response(body=out, status=201)
            else:
                return web.Response(status=204)
        except Exception as err:
            exception_type = type(err).__name__
            error_msg = f'Got an unexpected exception: {exception_type}, {err}'
            logging.error(error_msg)
            return web.Response(body=error_msg, status=500)

    async def __empty_messages_handler(self, _):
        return web.Response(status=204)

    async def __handle_generator(self):
        try:
            headers = {}

            if exists(AUTH_FILE_PATH) and isfile(AUTH_FILE_PATH):
                with open(AUTH_FILE_PATH, 'r') as file:
                    auth = file.read().replace('\n', '')
                    headers = {'Authorization': auth}

            async with ClientSession(headers=headers) as session:
                if isinstance(self.handler, GeneratorType):
                    for val in self.handler:
                        async with session.post(GENERATOR_STEP_SINK, data=val) as response:
                            if response.status > 299:
                                response_body = response.read()
                                logging.warn(
                                    'Message could not be processed', response.status, response_body)
                elif isinstance(self.handler, AsyncGeneratorType):
                    async for val in self.handler:
                        async with session.post(GENERATOR_STEP_SINK, data=val) as response:
                            if response.status > 299:
                                response_body = response.read()
                                logging.warn(
                                    'Message could not be processed', response.status, response_body)
        except Exception as err:
            logging.error('Generator function raised an error!')
            self.stop_request.set()
            raise err

    async def __start_background_tasks(self, app):
        app['generator_handler'] = create_task(self.__handle_generator())

    async def __cleanup_background_tasks(self, app):
        app['generator_handler'].cancel()
        await app['generator_handler']

    async def __on_shutdown(self, _):
        logging.debug('SDK Server shutting down.')

    def start(self, handler):
        self.handler = handler
        self.app = web.Application()
        self.app.add_routes([web.get('/ready', self.__ready_handler),
                            web.post('/messages', self.__messages_handler)])
        self.app.on_shutdown.append(self.__on_shutdown)
        logging.debug(
            f"Server starting at http://{HOST_NAME}:{SERVER_PORT} with pid {getpid()}")

        web.run_app(self.app, host=HOST_NAME,
                    port=SERVER_PORT, print=logging.debug)

    async def __start_generator(self, handler):
        handler_gen = handler()
        if not isinstance(handler_gen, GeneratorType) and not isinstance(handler_gen, AsyncGeneratorType):
            raise ValueError('Provided handler is not a generator function!')
        self.handler = handler_gen
        self.stop_request = asyncio.Event()
        self.app = web.Application()
        self.app.add_routes([web.get('/ready', self.__ready_handler),
                            web.post('/messages', self.__empty_messages_handler)])
        self.app.on_startup.append(self.__start_background_tasks)
        self.app.on_cleanup.append(self.__cleanup_background_tasks)
        self.app.on_shutdown.append(self.__on_shutdown)
        logging.debug(
            f"Generator Server starting at http://{HOST_NAME}:{SERVER_PORT} with pid {getpid()}")
        runner = web.AppRunner(self.app)
        await runner.setup()
        site = web.TCPSite(runner, HOST_NAME, SERVER_PORT)
        await site.start()
        logging.info(
            f"======== Running on {site.name} ========\n(Press CTRL+C to quit)")
        await self.stop_request.wait()
        logging.debug('Stop Request event received. Shutting down server.')
        await runner.cleanup()
        logging.debug('Stop Request event received. Cleanup done.')

    def start_generator(self, handler):
        logging.debug('Start generator called.')
        asyncio.run(self.__start_generator(handler))
        logging.debug('Start generator finished.')
