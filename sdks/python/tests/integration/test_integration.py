import subprocess
import os
import asyncio
import sys
import signal
import logging
import pytest
import pathlib
from aiohttp import ClientSession, web

if os.environ.get('PYTHONDEBUG'):
    logging.basicConfig(level=logging.info)


script_path = os.path.dirname(os.path.realpath(__file__))

# All test coroutines will be treated as marked.
pytestmark = pytest.mark.asyncio


class TestProcessHandler:
    async def asyncTearDown(self) -> None:
        if hasattr(self, 'process'):
            if not self.process.returncode:
                self.process.kill()
                await self.process.wait()
            del self.process

        if hasattr(self, 'runner'):
            await self.runner.cleanup()
            self.runner = None

    async def start_fixture(self, fixture, path_to_auth_token_file=None):
        env = {'PYTHONDEBUG': '1'}

        if path_to_auth_token_file:
            env['AUTH_FILE'] = path_to_auth_token_file

        self.process = await asyncio.create_subprocess_exec(sys.executable, f'{script_path}/../fixtures/{fixture}/app.py', stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=env)

        while True:
            line = (await self.process.stdout.readline()).decode('utf-8')
            logging.info('***** From background server *****: %s', line)
            if 'Running on' in line:
                break

        return self.process.stdout

    async def start_test_server(self, handler):
        app = web.Application()
        app.add_routes([web.post('/messages', handler)])
        self.runner = web.AppRunner(app)
        await self.runner.setup()
        site = web.TCPSite(self.runner, '0.0.0.0', 3569)
        await site.start()

        return self.runner

    async def test_default_step_handler(self):
        """
        Confirm that Sdk is able to run an Argo Dataflow Step handler fixture.
        """
        await self.start_fixture('default_step_handler')
        async with ClientSession() as session:
            async with session.get('http://localhost:8080/ready') as response:
                assert 204 == response.status

            async with session.post('http://localhost:8080/messages', data='Something') as response:
                assert 201 == response.status
                body = (await response.content.read()).decode('utf-8')
                assert 'Hi Something' == body
        await self.asyncTearDown()

    async def test_default_step_error_handler(self):
        """
        Confirm that Sdk is able to run an Argo Dataflow Step handler fixture that raises and error.
        """
        await self.start_fixture('default_step_error_handler')
        async with ClientSession() as session:

            async with session.post('http://localhost:8080/messages', data='Something') as response:
                assert 500 == response.status
                body = (await response.content.read()).decode('utf-8')
                assert 'Got an unexpected exception: ValueError, Some error' == body

        await self.asyncTearDown()

    async def test_default_step_async_handler(self):
        """
        Confirm that Sdk is able to run an Argo Dataflow Step asyncio handler fixture.
        """
        await self.start_fixture('default_step_async_handler')
        async with ClientSession() as session:

            async with session.post('http://localhost:8080/messages', data='Something') as response:
                assert 201 == response.status
                body = (await response.content.read()).decode('utf-8')
                assert 'Hi Something' == body

        await self.asyncTearDown()

    async def test_default_step_async_error_handler(self):
        """
        Confirm that Sdk is able to run an Argo Dataflow Step asyncio handler fixture that raises an error.
        """
        await self.start_fixture('default_step_async_error_handler')
        async with ClientSession() as session:

            async with session.post('http://localhost:8080/messages', data='Something') as response:
                assert 500 == response.status
                body = (await response.content.read()).decode('utf-8')
                assert 'Got an unexpected exception: AssertionError, ' == body

        await self.asyncTearDown()

    async def test_default_step_termination_handler(self):
        """
        Confirm that Sdk's server waits for requests to finish after SIGTERM is sent.
        """
        stdout = await self.start_fixture('default_step_termination_handler')

        async with ClientSession() as session:
            async def shutdown_background_server(proc):
                # This sleep makes sure that SIGTERM is sent after the request is started
                await asyncio.sleep(0.1)
                proc.send_signal(signal.SIGTERM)
                await self.process.wait()
                del self.process

            async def make_request(session):
                async with session.post('http://localhost:8080/messages', data='Something') as resp:
                    body = await resp.text()
                    return {'status': resp.status, 'body': body}

            result = await asyncio.gather(*[
                asyncio.ensure_future(make_request(session)),
                shutdown_background_server(self.process)
            ])

            assert 201 == result[0]['status']
            logging.info('***** From background server *****: ', await stdout.read())

        await self.asyncTearDown()

    async def test_generator_step_handler(self, tmp_path: pathlib.Path):
        """
        Confirm that Sdk is able to run an Argo Dataflow Generator Step.

        1. Create HTTP Test Server that listens on POST: http://0.0.0.0:3569/messages
        2. Pass inside an assert function accepting a request
        3. Start Argo_Dataflow_Sdk with a Generator Step fixture
        4. Assert that Sdk exposes a server responding to /ready & /messages endpoints.
        5. Assert that HTTP Test Server is getting requests with values from Sdk's Generator Step
        """
        auth_token = "someauthtoken"
        p = tmp_path / "authtoken"
        logging.info('Creating auth token at %s', p)
        p.write_text(auth_token)

        logging.info('creating test server, %s')
        requests = []
        bodies = []

        async def assert_incomming_msgs(request):
            nonlocal requests, bodies
            body = await request.text()
            requests.append(request)
            bodies.append(body)
            return web.Response(status=200)

        await self.start_test_server(assert_incomming_msgs),

        logging.info('test server created')

        logging.info('creating fixture')
        await self.start_fixture('generator_step_handler', path_to_auth_token_file=p)
        logging.info('fixture created')

        async with ClientSession() as session:
            logging.info('about to send first request')
            async with session.get('http://localhost:8080/ready') as response:
                logging.info('got response from first request %i',
                             response.status)
                assert 204 == response.status

            logging.info('about to send second request')
            async with session.post('http://localhost:8080/messages', data='Something') as response:
                logging.info(
                    'got response from second request %i', response.status)
                assert 204 == response.status
                body = await response.text()
                logging.info('got body from second request %s', body)
                assert '' == body

        i = 0
        for body in bodies:
            assert f"Some Value {i}" == body
            i = i + 1

        for request in requests:
            assert auth_token == request.headers['Authorization']

        await self.asyncTearDown()

    async def test_generator_step_error_handler(self, tmp_path):
        """
        Confirm that Sdk is able to run an Argo Dataflow Generator Step that raises an error.
        This case should stop the whole Sdk server and kill the process with exit code 1.
        """
        auth_token = "someauthtoken"
        p = tmp_path / "authtoken"
        logging.info('Creating auth token at %s', p)
        p.write_text(auth_token)

        logging.info('creating test server')
        requests = []
        bodies = []

        async def assert_incomming_msgs(request):
            nonlocal requests, bodies
            body = await request.text()
            requests.append(request)
            bodies.append(body)
            return web.Response(status=200)

        await self.start_test_server(assert_incomming_msgs),
        logging.info('test server created')

        logging.info('creating fixture')
        await self.start_fixture('generator_step_error_handler', path_to_auth_token_file=p)
        logging.info('fixture created')

        while not self.process.returncode:
            logging.info('***** Background server logs *****: %s - %s', await self.process.stdout.readline(), self.process.returncode)

        logging.info('background process done, %s', self.process.returncode)
        i = 0
        for body in bodies:
            assert f"Some Value {i}" == body
            i = i + 1

        for request in requests:
            assert auth_token == request.headers['Authorization']

        assert 1 == self.process.returncode

        await self.asyncTearDown()

    async def test_generator_step_async_handler(self, tmp_path):
        """
        Confirm that Sdk is able to run an Argo Dataflow Generator Step asyncio handler.

        1. Create HTTP Test Server that listens on POST: http://0.0.0.0:3569/messages
        2. Pass inside an assert function accepting a request
        3. Start Argo_Dataflow_Sdk with a Generator Step fixture
        4. Assert that Sdk exposes a server responding to /ready & /messages endpoints.
        5. Assert that HTTP Test Server is getting requests with values from Sdk's Generator Step
        """
        auth_token = "someauthtoken"
        p = tmp_path / "authtoken"
        logging.info('Creating auth token at %s', p)
        p.write_text(auth_token)

        logging.info('creating test server')
        requests = []
        bodies = []

        async def assert_incomming_msgs(request):
            nonlocal requests, bodies
            body = await request.text()
            requests.append(request)
            bodies.append(body)
            return web.Response(status=200)

        await self.start_test_server(assert_incomming_msgs),

        logging.info('test server created')

        logging.info('creating fixture')
        await self.start_fixture('generator_step_async_handler', path_to_auth_token_file=p)
        logging.info('fixture created')

        async with ClientSession() as session:
            logging.info('about to send first request')
            async with session.get('http://localhost:8080/ready') as response:
                logging.info('got response from first request %i',
                             response.status)
                assert 204 == response.status

            logging.info('about to send second request')
            async with session.post('http://localhost:8080/messages', data='Something') as response:
                logging.info(
                    'got response from second request %i', response.status)
                assert 204 == response.status
                body = await response.text()
                logging.info('got body from second request %s', body)
                assert '' == body

        i = 0
        for body in bodies:
            assert f"Some Value {i}" == body
            i = i + 1

        for request in requests:
            assert auth_token == request.headers['Authorization']

        await self.asyncTearDown()

    async def test_generator_step_async_error_handler(self, tmp_path):
        """
        Confirm that Sdk is able to run an Argo Dataflow Generator Step asyncio handler that raises an error.
        This case should stop the whole Sdk server and kill the process with exit code 1.
        """
        auth_token = "someauthtoken"
        p = tmp_path / "authtoken"
        logging.info('Creating auth token at %s', p)
        p.write_text(auth_token)

        logging.info('creating test server')
        requests = []
        bodies = []

        async def assert_incomming_msgs(request):
            nonlocal requests, bodies
            body = await request.text()
            requests.append(request)
            bodies.append(body)
            return web.Response(status=200)

        await self.start_test_server(assert_incomming_msgs),
        logging.info('test server created')

        logging.info('creating fixture')
        await self.start_fixture('generator_step_async_error_handler', path_to_auth_token_file=p)
        logging.info('fixture created')

        while not self.process.returncode:
            logging.info('***** Background server logs *****: %s - %s', await self.process.stdout.readline(), self.process.returncode)

        logging.info('background process done, %s', self.process.returncode)
        i = 0
        for body in bodies:
            assert f"Some Value {i}" == body
            i = i + 1

        for request in requests:
            assert auth_token == request.headers['Authorization']

        assert 1 == self.process.returncode

        await self.asyncTearDown()
