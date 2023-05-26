import base64
import nats
import os
import sys
import asyncio
import socket
def get_container_ip():
    # 获取容器内部的IP地址
    ip = socket.gethostbyname(socket.gethostname())
    return ip


print(get_container_ip())

async def handle(req):
    # 将请求数据编码为base64格式

    nc = await nats.connect(os.environ['NATS_ADDRESS'])
    req_data = base64.b64decode(req.encode('utf-8'))
    print(sys.getsizeof(req_data))
    # 连接到NATS服务器
    sub = os.environ['NATS_SUBJECT']
    await nc.publish(sub, payload=req_data)
        # 关闭NATS连接
    await nc.drain()
    # 返回响应
    return 'Request received and queued for processing.'





