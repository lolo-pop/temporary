import json
import os

# 加载JSON文件到Python对象
with open('/home/rongch05/openfaas/training/results_by_appID/TFT.json', 'r') as f:
    data = json.load(f)

# 切割数据并写入多个小文件
print(type(data))
chunk_size = 1000 # 每个小文件的数据量
output_dir = 'split_results_by_appID/TFT'  # 小文件的输出目录
if not os.path.exists(output_dir):
    os.makedirs(output_dir)

chunks = []
tmp = list(data.items())
print(len(data), data.keys())

# 'agg_metrics', 'item_metrics'
for key in data.keys():
    print(type(data[key]))
    for i in range(0, len(data[key]), chunk_size):
        chunk = data[key][i:i+chunk_size]
        print(len(chunk))
        chunks.append(chunk)

    for i, chunk in enumerate(chunks):
        filename = f'{output_dir}/{key}_chunk_{i}.json'
        print(filename)
        with open(filename, 'w') as f:
            f.write(json.dumps(chunk, indent=4, sort_keys=True))