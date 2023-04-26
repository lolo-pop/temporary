"""
训练SOTA算法的预测模型
- 算法list：Estimators
- 输出：Path(model_path)
"""

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import json
import os
from gluonts.dataset.common import ListDataset
from gluonts.dataset.field_names import FieldName
from gluonts.mx import DeepAREstimator, TransformerEstimator, SimpleFeedForwardEstimator, \
    TemporalFusionTransformerEstimator, DeepFactorEstimator, MQRNNEstimator
from gluonts.mx import Trainer
from gluonts.evaluation import backtest_metrics, make_evaluation_predictions, Evaluator
from pathlib import Path
from gluonts.transform import AddObservedValuesIndicator
from concurrent.futures import ThreadPoolExecutor
import config as config

    
    

def train_model(train_config, split_by_ratio):
    """
    针对一个数据集，训练模型并保存到指定位置
    train_config: 一个dict，包括数据集位置、数据集元数据等
    """
    # 数据预处理
    total_target = []
    total_start_time = []

    # 根据数据集格式处理
    if train_config["name"] in ["rtn", "measurement"]:
        for file in train_config["data_path"].iterdir():
            cur_target = []
            cur_start_time = -1
            line_count = 0
            with open(file, 'r', encoding='utf-8') as f:
                for line in f.readlines():
                    line_count += 1
                    if line_count == 1:
                        cur_start_time = json.loads(line)["timestamp(UTC)"]
                    # 如果不是依据比例划分，则在指定位置截取数据集
                    if not split_by_ratio and line_count > train_config["test_ds_length"]:
                        total_target.append(cur_target)
                        total_start_time.append(pd.to_datetime(cur_start_time, unit="s"))
                        break
                    avg_rtt = json.loads(line)["avg_rtt"]
                    if avg_rtt > 0:
                        cur_target.append(avg_rtt)
                    else:
                        cur_target.append(np.NaN)
            if split_by_ratio:
                total_target.append(cur_target)
                total_start_time.append(pd.to_datetime(cur_start_time, unit="s"))

    if split_by_ratio:
        train_ds = ListDataset(
            data_iter=[
                {
                    FieldName.TARGET: total_target[i][:int(train_config["train_ds_ratio"]*len(total_target[i]))],
                    FieldName.START: total_start_time[i]
                }
                for i in range(0, len(total_target))
            ],
            freq=train_config["frequency"]
        )
    else:
        train_ds = ListDataset(
            data_iter = [
                {
                    FieldName.TARGET: total_target[i][:train_config["train_ds_length"]],
                    FieldName.START: total_start_time[i]
                }
                for i in range(0,len(total_target))
            ],
            freq=train_config["frequency"]
        )


    # 训练模型, 每个预测长度一个
    for prediction_length in train_config["prediction_length"]:
        Estimators = {
            "DeepAR": DeepAREstimator(
                    prediction_length=prediction_length,
                    freq=train_config["frequency"],
                    context_length = 10
                ),
            "Transformer": TransformerEstimator(
                    prediction_length=prediction_length,
                    freq=train_config["frequency"],
                    prediction_length = 10
                ),
            # TFT基于quantile regression进行分布预测，通过参数num_outputs指定quantile数量
            #     quantile_list = sum(
            #         ([i / 10, 1.0 - i / 10] for i in range(1, (num_outputs + 1) // 2)),
            #         [0.5]
            #     ),
            #     e.g., 当num_outputs=3时，quantile_list = [0.1, 0.5, 0.9]
            #     所以num_outputs最大为9，且只能为奇数
            "TFT": TemporalFusionTransformerEstimator(
                    prediction_length=prediction_length,
                    freq=train_config["frequency"],
                    num_outputs=9, # quantile regression，用于指定quantile数量
                    context_length = 10
                ),
            "FeedforwardNN": SimpleFeedForwardEstimator(
                    prediction_length=prediction_length,
                    context_length = 10
                ),
            "DeepFactor": DeepFactorEstimator(
                    prediction_length=prediction_length,
                    freq=train_config["frequency"],
                    context_length = 10
                ),
            "MQRNN": MQRNNEstimator(
                    prediction_length=prediction_length,
                    freq=train_config["frequency"],
                    quantiles=config.evaluation_quantile, 
                    context_length = 10
                )
        }

        def train_process(model_name):
            print(f"start training {model_name} in {prediction_length} prediction length")
            Estimator = Estimators[model_name]
            model_output_path = train_config["model_path"] / f"{prediction_length}_prediction_length" / f"{model_name}"
            predictor = Estimator.train(train_ds)
            os.makedirs(model_output_path, exist_ok=True)
            predictor.serialize(model_output_path)

        # 使用多线程并行训练
        pool = ThreadPoolExecutor(max_workers=10)
        pool.map(train_process, config.train_model)


if __name__ == "__main__":
    split_by_ratio = True # 根据比例切分训练集和测试集的比例
    for dataset_meta_data in config.dataset_meta_data:
        if dataset_meta_data["name"] in config.train_dataset:
            train_model(dataset_meta_data, split_by_ratio)

