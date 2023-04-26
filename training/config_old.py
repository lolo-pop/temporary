import pandas as pd
from pathlib import Path

evaluation_quantile = [i/20 for i in range(1,20)] + [0.99]
prediction_length = [20, 40, 60, 2*60, 3*60, 5*60, 10*60, 20*60, 30*60] # seconds
baseline_model = ["DeepAR", "Transformer", "TFT", "FeedforwardNN", "DeepFactor", "MQRNN"]
train_model = ["TFT"] # 用于指定这次要训练的模型
train_dataset = ["rtn"] # 用于指定这次要训练的数据集
evaluate_model = ["TFT", "DeepAR", "Transformer"] # 用于指定这次要测试的模型
evaluate_dataset = ["rtn"] # 用于指定这次要测试的数据集

dataset_meta_data = [
    {
        "name" : "rtn",
        "data_path" : Path("../../Dataset/propritary/processed/rtn_sorted/"),
        "model_path" : Path("../model/rtn/"),
        "evaluate_result_path": Path("../results/rtn/"),
        "point_result_path": Path("../point results/rtn/"),
        "train_ds_length" : 24 * 60 * 60 // 20,    # 这个是分割数据集 训练集
        "test_ds_length" : (24+2) * 60 * 60 // 20, # 测试集
        "train_ds_ratio" : 0.9,
        "prediction_length" : [i // 20 for i in prediction_length],   # 预测下一个时间窗口取值 1 
        "frequency" : "20S"
    },
    {
        "name" : "measurement",
        "data_path" : Path("../../Dataset/measurement/"),
        "model_path" : Path("../model/measurement/"),
        "evaluate_result_path": Path("../results/measurement/"),
        "point_result_path": Path("../point results/measurement/"),
        "train_ds_length" : 10 * 24 * 60 * 60 // 2,
        "test_ds_length" : (10 * 24 * 60 * 60 // 2) + (1 * 24 * 60 * 60 // 2),
        "prediction_length" : [i * 60 // 2 for i in prediction_length],
        "frequency" : "2S"
    },


]