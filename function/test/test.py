import requests
import base64
import sys

if __name__ == '__main__':
    # 读取图片文件并编码为base64格式
    with open('/home/rongch05/openfaas/inception.png', 'rb') as f:
        image_data = base64.b64encode(f.read()).decode('utf-8')
        print(sys.getsizeof(base64.b64decode(image_data.encode('utf-8'))))
    # 发送HTTP POST请求并将图片数据作为请求体
    response = requests.post('http://127.0.0.1:8080/function/nats-test/', data=image_data)
    response = requests.post('http://127.0.0.1:8080/function/nats-test/', data=image_data)
    
    
    # 打印响应内容
    print(response.text)
