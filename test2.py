acc_dict = {
  0: [0, 18],
	1: [18, 22],
	2: [22, 24],
	3: [24, 28],
	4: [28, 30],
	5: [30, 32],
}

a =24.090
level = -1
for key, value in acc_dict.items():
  if a < value[1] and a >= value[0]:
    level = key

print(level)