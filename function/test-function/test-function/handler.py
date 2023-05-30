import logging
import os
import time
import base64
import io
import ctypes
import json
from urllib.request import Request, urlopen
from PIL import Image
import socket

r_size = int(os.environ['R_SIZE'])
c_size = int(os.environ['C_SIZE'])

logging.basicConfig(
    level=logging.DEBUG,  # 日志级别
    format='%(asctime)s %(levelname)s %(message)s',  # 日志格式
)

dispatch_url = "http://10.244.0.24:5000/sendImage"

class image:
    def __init__(self, name, from_, data):
        self.name = name
        self.from_ = from_
        self.data = data

    def to_json(self):
        return json.dumps(self.__dict__)

def get_container_ip():
    hostname = socket.gethostname()
    ip = socket.gethostbyname(hostname)
    return ip

def send_image(i, f):
    tmp = f"image{i}.png"
    image_data = {
        "name": tmp,
        "from": get_container_ip(),
        "data": tmp
    }
    json_data = json.dumps(image_data)
    req = Request(dispatch_url, json_data.encode(), {"Content-Type": "application/json"}, method='POST')
    try:
        with urlopen(req) as resp:
            print(f"Image sent with status: {resp.status}")
    except Exception as e:
        print(f"Error sending image: {e}")

def sender(num1, num2):
    t1 = time.time()
    lib = ctypes.CDLL('/home/app/function/sender.so')
    lib.Test(num1, num2)
    t2 = time.time()
    logging.info(f"sender image time:"+str(t2-t1))

def handle(req):
    # 将请求数据编码为base64格式

    # logging.debug('Debug message')
    # logging.info('Info message')
    # logging.warning('Warning message')
    # logging.error('Error message')
    # logging.critical('Critical message')
    t1 = time.time()
    # req_data = base64.b64decode(req.encode('utf-8'))
    # image_stream = io.BytesIO(req_data)
    image_stream = "/home/app/function/inception.png"
    img = Image.open(image_stream)
    img = img.convert('RGB')
    resized_img = img.resize((r_size, c_size), Image.LANCZOS)
    t2 = time.time()
    resized_img.save("1.jpg")
    t3 = time.time()
    logging.info(f"resize time:"+str(t2 - t1))
    logging.info(f"save image time:"+str(t3-t2))
    sender(1, 1)
    # send_image(1, 1)
    return 'Request received.'





