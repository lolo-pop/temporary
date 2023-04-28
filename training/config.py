import pandas as pd
from pathlib import Path

evaluation_quantile = [0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9]
baseline_model = ["DeepAR", "Transformer", "TFT", "FeedforwardNN", "DeepFactor", "MQRNN"]
train_model = ["DeepAR", "Transformer", "FeedforwardNN", "DeepFactor", "MQRNN"] # 用于指定这次要训练的模型
# train_model = ["TFT"]
evaluate_model = ["TFT"] # 用于指定这次要测试的模型


dataset_meta_data = {
        "data_path" : Path("/home/rongch05/openfaas/AzurePublicDataset/invocationTraceByAppID.csv"),
        "model_path" : Path("predict_models_by_appID/"),
        "evaluate_result_path": Path("results_by_appID/"),
        "point_result_path": Path("point_results_by_appID/"),
        "train_ds_ratio" : 0.98,
        "prediction_length" : 1,   # 预测下一个时间窗口取值 1 
        "frequency" : "30s"
    }

