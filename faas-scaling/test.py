import csv

# 打开csv文件并读取数据
with open('profiling.csv', 'r') as csvfile:
    reader = csv.reader(csvfile)
    rows = [row for row in reader]

# 将第一列的所有值减1
for i in range(len(rows)):
  if i != 0:
    rows[i][5] = str(float(rows[i][5]) + 0.15)

# 将修改后的数据写回原始文件中
with open('your_file.csv', 'w', newline='') as csvfile:
    writer = csv.writer(csvfile)
    writer.writerows(rows)
