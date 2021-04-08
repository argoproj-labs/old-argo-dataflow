import requests
from http.server import BaseHTTPRequestHandler, HTTPServer

from handler import handler


class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):  # GET /ready
        self.send_response(200)
        self.end_headers()

    def do_POST(self):  # POST /messages
        len = int(self.headers.get('Content-Length'))
        msg = self.rfile.read(len)
        for msg in handler(msg):
            resp = requests.post('http://localhost:3569/messages', data=msg)
            if resp.status_code != 200:
                raise Exception(resp.status_code)
        self.send_response(200)
        self.end_headers()


if __name__ == "__main__":
    webServer = HTTPServer(("localhost", 8080), MyServer)

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()
