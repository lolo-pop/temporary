import requests
import base64
import sys
import time 
import threading


def request_thread():
    response = requests.post('http://127.0.0.1:8080/function/test-function-0/', data="")
    # 处理响应结果
    print(response.text)
# 启动5个线程，每个线程执行一次请求
t1 = time.time()
request_thread()
t2 = time.time()
print(t2-t1)
# for i in range(3):
#     t = threading.Thread(target=request_thread)
#     t.start()