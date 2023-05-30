import ctypes
import time

t1 = time.time()
lib = ctypes.CDLL('./sender.so')



name = "test"
imageName = "test.png"

lib.Test(0, 1)
t2 = time.time()

print(t2-t1)
