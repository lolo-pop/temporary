import requests
import base64
import nats
import os

def handle(req):
    # 将请求数据编码为base64格式
    req_data = base64.b64encode(req.content).decode()
    print(os.environ['NATS_ADDRESS'])
    print(os.environ['NATS_SUBJECT'])
    # 连接到NATS服务器
    nc = nats.connect(servers=[os.environ['NATS_ADDRESS']])
    subject = os.environ['NATS_SUBJECT']
    # 将请求数据发送到NATS队列
    nc.publish(subject, req_data.encode())
    # 关闭NATS连接
    nc.close()

    # 返回响应
    return 'Request received and queued for processing.'