import requests
import base64
import sys
import time 
if __name__ == '__main__':
    # 读取图片文件并编码为base64格式
    t1 = time.time()
    response = requests.post('http://127.0.0.1:8080/function/test-function-0/', data="")
    t2 = time.time()
    response = requests.post('http://127.0.0.1:8080/function/test-function-0/', data="")
    t3 = time.time()
    print(t2-t1)
    print(t3-t2)
    # 打印响应内容
    print(response.text)