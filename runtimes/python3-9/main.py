from http.server import BaseHTTPRequestHandler, HTTPServer
import requests

class MyServer(BaseHTTPRequestHandler):
    def do_POST(self):
        content_len = int(self.headers.get('Content-Length'))
        post_body = self.rfile.read(content_len)
        msg = handler(post_body)
        if requests.post('http://localhost:3569/messages', data = msg).statusCode != 200:
            raise Exception("error")
        self.send_response(200)
        self.end_headers()
if __name__ == "__main__":
    webServer = HTTPServer(("localhost", 8080), MyServer)

    try:
        webServer.serve_forever()
    except KeyboardInterrupt:
        pass

    webServer.server_close()