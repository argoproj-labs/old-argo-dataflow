import requests
from http.server import BaseHTTPRequestHandler, HTTPServer

from handler import handler


class MyServer(BaseHTTPRequestHandler):
    def do_GET(self):  # GET /ready
        self.send_response(204)
        self.end_headers()

    def do_POST(self):  # POST /messages
        len = int(self.headers.get('Content-Length'))
        msg = self.rfile.read(len)
        out = handler(msg)
        if out:
            self.send_response(201)
            self.end_headers()
            self.wfile.write(out)
        else:
            self.send_response(204)
            self.end_headers()


if __name__ == "__main__":
    webServer = HTTPServer(("localhost", 8080), MyServer)

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()
