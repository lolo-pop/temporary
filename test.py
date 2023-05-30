import time
from PIL import Image

image_stream = "/home/rongch05/openfaas/function/service/service/image2.jpg"
img = Image.open(image_stream)
img = img.convert('RGB')
resized_img = img.resize((600, 600), Image.LANCZOS)
t2 = time.time()
resized_img.save("image.jpg")
t3 = time.time()