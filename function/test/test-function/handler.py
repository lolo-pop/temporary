import logging
import os
import time
import base64
import io
from PIL import Image


r_size = int(os.environ['R_SIZE'])
c_size = int(os.environ['C_SIZE'])

def handle(req):
    # 将请求数据编码为base64格式
    logging.basicConfig(
        level=logging.DEBUG,  # 日志级别
        format='%(asctime)s %(levelname)s %(message)s',  # 日志格式
    )
    # logging.debug('Debug message')
    # logging.info('Info message')
    # logging.warning('Warning message')
    # logging.error('Error message')
    # logging.critical('Critical message')
    t1 = time.time()
    req_data = base64.b64decode(req.encode('utf-8'))
    image_stream = io.BytesIO(req_data)
    # image_path = "/home/app/function/image.jpg"
    img = Image.open(image_stream)
    resized_img = img.resize((r_size, c_size), Image.LANCZOS)
    t2 = time.time()
    resized_img.save("1.jpg")
    t3 = time.time()
    logging.info("resize time:", t2 - t1)
    logging.info("save image time:", t3-t2)
    return 'Request received.'





