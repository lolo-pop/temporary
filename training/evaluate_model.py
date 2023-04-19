"""
    1、基于训练的模型进行预测，输出预测准确度结果
    2、将预测的point forecast结果（即p50 quantile）保存，作为global model的结果
"""

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
import json
import os
from gluonts.dataset.common import ListDataset
from gluonts.dataset.field_names import FieldName
from gluonts.mx import DeepAREstimator, TransformerEstimator, SimpleFeedForwardEstimator, \
    TemporalFusionTransformerEstimator
from gluonts.mx import Trainer
from gluonts.evaluation import backtest_metrics, make_evaluation_predictions, Evaluator
from gluonts.model.predictor import Predictor
from pathlib import Path
import config
import datetime


def evaluate_model(evaluate_config, split_by_ratio):
    """
        针对一个数据集，测试模型并保存结果到指定位置
        evaluate_config: 一个dict，包括数据集位置、模型位置等
    """
    # 数据预处理
    total_target = []
    total_start_time = []
    total_link_name = []

    # 根据数据集格式处理
    if evaluate_config["name"] in ["rtn", "measurement"]:
        for file in evaluate_config["data_path"].iterdir():
            cur_target = []
            cur_start_time = -1
            line_count = 0
            with open(file, 'r', encoding='utf-8') as f:
                for line in f.readlines():
                    line_count += 1
                    if line_count == 1:
                        cur_start_time = json.loads(line)["timestamp(UTC)"]
                    # 如果不是依据比例划分，则在指定位置截取数据集
                    if not split_by_ratio and line_count > evaluate_config["test_ds_length"]:
                        total_target.append(cur_target)
                        total_start_time.append(pd.to_datetime(cur_start_time, unit="s"))
                        total_link_name.append(file.parts[-1])
                        break
                    avg_rtt = json.loads(line)["avg_rtt"]
                    if avg_rtt > 0:
                        cur_target.append(avg_rtt)
                    else:
                        cur_target.append(np.NaN)
            if split_by_ratio:
                total_target.append(cur_target)
                total_start_time.append(pd.to_datetime(cur_start_time, unit="s"))
                total_link_name.append(file.parts[-1])

    # 对每个预测长度进行评估，
    for prediction_length in evaluate_config["prediction_length"]:
        # 相关路径
        model_path = evaluate_config["model_path"] / f"{prediction_length}_prediction_length"
        result_path = evaluate_config["evaluate_result_path"] / f"{prediction_length}_prediction_length"
        point_result_path = evaluate_config["point_result_path"] / f"{prediction_length}_prediction_length"
        os.makedirs(result_path, exist_ok=True)
        os.makedirs(point_result_path, exist_ok=True)

        for model in config.evaluate_model:
            # 用于记录这个模型的结果,分链路记录
            global_model_result = {
                link: {
                    "point_forecast_result": [],
                    "observation": []
                }
                for link in total_link_name
            }
            evaluation_result = {
                "agg_metrics": [],
                "item_metrics": []
            }

            # 根据预测长度，滑动窗口构造测试集（预测每个序列末尾prediction_length长度的部分）
            # 但是这样会把一个series扩为多个，memory会爆：因此逐个series进行evaluation，而非一次丢进去
            # 每个循环测试一个time series
            for i, target in enumerate(total_target):
                test_pairs = []
                if split_by_ratio:
                    test_data_length = int(evaluate_config["train_ds_ratio"] * len(target)) + prediction_length
                else:
                    test_data_length = evaluate_config["train_ds_length"] + prediction_length
                while test_data_length <= len(target):
                    test_pairs.append((total_start_time[i], target[:test_data_length]))
                    test_data_length += prediction_length
                test_ds = ListDataset(
                    data_iter=[
                        {
                            FieldName.TARGET: target,
                            FieldName.START: start
                        }
                        for start, target in test_pairs
                    ],
                    freq=evaluate_config["frequency"]
                )

                # 加载训练好的模型
                predictor = Predictor.deserialize(model_path / model)
                forecast_it, ts_it = make_evaluation_predictions(
                    dataset=test_ds,  # test dataset
                    predictor=predictor,  # predictor
                    num_samples=500,  # number of sample paths we want for evaluation
                )

                forecasts = list(forecast_it)
                tss = list(ts_it)
                link_name = total_link_name[i]

                # 记录点预测结果，作为global model的输出
                for j, probabilistic_f in enumerate(forecasts):
                    global_model_result[link_name]["point_forecast_result"].append(probabilistic_f.quantile(0.5).tolist())
                    global_model_result[link_name]["observation"].append(
                        np.array(tss[j][-prediction_length:]).reshape(-1).tolist())

                # evaluate概率预测的结果
                # num_workers默认为CPU数，测试设备为一个16核的cpu
                evaluator = Evaluator(quantiles=config.evaluation_quantile, num_workers=10)
                agg_metrics, item_metrics = evaluator(tss, forecasts)

                evaluation_result["agg_metrics"].append(agg_metrics)
                evaluation_result["item_metrics"].append(item_metrics.to_json(orient='records'))

            with open(result_path / f"{model}.json", "w") as f:
                f.write(json.dumps(evaluation_result, indent=4, sort_keys=True))
            with open(point_result_path / f"{model}.json", "w") as f:
                f.write(json.dumps(global_model_result, indent=4, sort_keys=True))


if __name__ == "__main__":
    split_by_ratio = True  # 根据比例切分训练集和测试集的比例
    for dataset_meta_data in config.dataset_meta_data:
        if dataset_meta_data["name"] in config.evaluate_dataset:
            evaluate_model(dataset_meta_data, split_by_ratio)
