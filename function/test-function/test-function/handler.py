import logging
import os
import time
import base64
import io
import ctypes
import json
import redis
import uuid

from urllib.request import Request, urlopen
from PIL import Image
import socket

r_size = int(os.environ['R_SIZE'])
c_size = int(os.environ['C_SIZE'])

logging.basicConfig(
    level=logging.DEBUG,  # 日志级别
    format='%(asctime)s %(levelname)s %(message)s',  # 日志格式
)

dispatch_url = "http://faas-dispatch.openfaas.svc.cluster.local:5000/sendImage"


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

def send_image(name, ip, data):
    
    image_data = {
        "name": name,
        "from": ip,
        "data": data
    }
    json_data = json.dumps(image_data)
    req = Request(dispatch_url, json_data.encode(), {"Content-Type": "application/json"}, method='POST')
    try:
        with urlopen(req) as resp:
            print(f"Image sent with status: {resp.status}")
    except Exception as e:
        print(f"Error sending image: {e}")
    
    redis_host = "faas-redis-master.openfaas.svc.cluster.local"
    redis_port = 6379
    redis_db = 0
    redis_password = "Y7MkRCBORP"
    r = redis.Redis(host=redis_host, port=redis_port, db=redis_db, password=redis_password)
    while True:
        if r.exists(name):
            value = r.get(name)
            print("received "+name+" results "+str(value))
            break
        time.sleep(0.001)
    
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
    data = "image1.jpg"
    resized_img.save(data)
    t3 = time.time()
    logging.info(f"resize time:"+str(t2 - t1))
    logging.info(f"save image time:"+str(t3-t2))
    # sender(1, 1)
    t4 = time.time()
    uuid_name = uuid.uuid1()
    name = str(uuid_name)
    ip = get_container_ip()
    send_image(name, ip, data)
    t5 = time.time()
    logging.info(f"send image time:"+str(t5-t4))
    logging.info(f"total time:"+str(t5-t1))
    return 'Request received.'





