import requests
import base64
import sys

if __name__ == '__main__':
    # 读取图片文件并编码为base64格式
    # 发送HTTP POST请求并将图片数据作为请求体
    response = requests.post('http://127.0.0.1:8080/function/test/', data="dasda")
    # response = requests.post('http://127.0.0.1:8080/function/test/', data="dassd")
    
    
    # 打印响应内容
    print(response.text)