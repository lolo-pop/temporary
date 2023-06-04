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
accuracy = float(os.environ["accuracy"])
latency = float(os.environ["latency"])
functionName = os.environ["name"]


#dispatch_url = "http://faas-dispatch.openfaas.svc.cluster.local:5000/sendImage"

acc_dict = {
    0: [0, 18],
	1: [18, 22],
	2: [22, 24],
	3: [24, 28],
	4: [28, 30],
	5: [30, 32],
}
level = -1

for key, value in acc_dict.items():
  if accuracy < value[1] and accuracy >= value[0]:
    level = key
    break

redis_host = "faas-redis-master.openfaas.svc.cluster.local"
redis_port = 6379
redis_db = 0
redis_password = "Y7MkRCBORP"
dispatch_url = f"http://faas-dispatch-{level}.openfaas.svc.cluster.local:5000/sendImage"

slo = {
    "accuracy" : accuracy,
    "latency" : latency
}
r = redis.Redis(host=redis_host, port=redis_port, db=redis_db, password=redis_password)
json_data = json.dumps(slo)
r.set(functionName, json_data)

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
    
    r = redis.Redis(host=redis_host, port=redis_port, db=redis_db, password=redis_password)
    while True:
        if r.exists(name):
            value = r.get(name)
            print("received "+name+" results "+str(value))
            break
        time.sleep(0.001)
    

    
    
def handle(event, context):
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
    # sender(1, 1)
    t4 = time.time()
    uuid_name = uuid.uuid1()
    name = str(uuid_name)
    ip = get_container_ip()
    send_image(name, ip, data)
    t5 = time.time()
    json_data = f"resize time: {t2 - t1}\nsave image time:{t3-t2}\nsend image time:{t5-t4}\ntotal time:{t5-t1}"
    return {
        "statusCode": 200,
        "body": json_data
    }
