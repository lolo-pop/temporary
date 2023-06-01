import requests
import base64
import sys
import time 
import threading



while True:
  t1 = time.time()
  response = requests.post('http://127.0.0.1:8080/function/test-function-0/', data="")
  print(response.text)
  t2 = time.time()
  print(t2-t1)
