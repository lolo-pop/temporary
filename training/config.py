import pandas as pd
from pathlib import Path

evaluation_quantile = [i/20 for i in range(1,20)] + [0.99]
prediction_length = [20, 40, 60, 2*60, 3*60, 5*60, 10*60, 20*60, 30*60] # seconds
baseline_model = ["DeepAR", "Transformer", "TFT", "FeedforwardNN", "DeepFactor", "MQRNN"]
train_model = ["DeepAR", "Transformer", "TFT", "FeedforwardNN", "DeepFactor", "MQRNN"] # 用于指定这次要训练的模型
evaluate_model = ["TFT"] # 用于指定这次要测试的模型


dataset_meta_data = {
        "data_path" : Path("/home/rongch05/openfaas/AzurePublicDataset/invocation.csv"),
        "model_path" : Path("predict_models/"),
        "evaluate_result_path": Path("results/"),
        "point_result_path": Path("point results/"),
        "train_ds_ratio" : 0.9,
        "prediction_length" : 1,   # 预测下一个时间窗口取值 1 
        "frequency" : "30s"
    }

