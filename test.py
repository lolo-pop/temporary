import numpy as np
import tensorflow as tf
import time
from PIL import Image
import os
import json


t1 =time.time()
image_path = "/home/rongch05/openfaas/function/service-2/service/image.jpg"
im = Image.open(image_path)
t2 =time.time()
print(t2-t1)
t3 = time.time()
image_np = np.array(im)
ls = []
ls.append(image_np)
c = np.array(ls)
input_tensor = tf.convert_to_tensor(c, dtype=tf.uint8)
print(input_tensor.shape)
t4 = time.time()
print(t4-t3)


