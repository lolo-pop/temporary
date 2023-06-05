import requests
import base64
import sys
import time
if __name__ == '__main__':
    # 读取图片文件并编码为base64格式
    # 发送HTTP POST请求并将图片数据作为请求体
    t1 = time.time()
    response = requests.post('http://127.0.0.1:8080/function/test-1/', data="")
    # response = requests.post('http://127.0.0.1:8080/function/test/', data="dassd")
    print(response.text)    
    t2 = time.time()
    print(t2-t1)
    