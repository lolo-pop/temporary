import json
import random
import socket
import time
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Thread
from urllib.request import Request, urlopen

dispatch_url = "http://10.244.0.18:5000/sendImage"


class ResultHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path != "/sendResult":
            self.send_response(405)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"Method not allowed")
            return

        content_length = int(self.headers.get("Content-Length", 0))
        data = self.rfile.read(content_length)
        try:
            result = json.loads(data)
        except json.JSONDecodeError:
            self.send_response(400)
            self.send_header("Content-type", "text/plain")
            self.end_headers()
            self.wfile.write(b"Bad request")
            return

        self.send_response(200)
        self.end_headers()

        self.server.results.append(result)
        self.server.shutdown()


def get_results():
    server = HTTPServer(("", 8084), ResultHandler)
    server.results = []
    server.serve_forever()
    return server.results


def test(i, f):
    send_image(0, 1)
    results_thread = Thread(target=get_results, daemon=True)
    results_thread.start()
    results_thread.join()
    result = results_thread.result
    print("test", result)


def main():
    random.seed(time.time())
    bs = 5
    test(0, bs)


def get_container_ip():
    hostname = socket.gethostname()
    ip = socket.gethostbyname(hostname)
    return ip


class Image:
    def __init__(self, name, from_, data):
        self.name = name
        self.from_ = from_
        self.data = data

    def to_json(self):
        return json.dumps(self.__dict__)


def send_image(i, f):
    tmp = f"image{i}.png"
    image_data = Image(tmp, get_container_ip(), tmp)
    json_data = image_data.to_json()
    req = Request(dispatch_url, json_data.encode(), {"Content-Type": "application/json"})
    try:
        with urlopen(req) as resp:
            print(f"Image sent with status: {resp.status}")
    except Exception as e:
        print(f"Error sending image: {e}")
