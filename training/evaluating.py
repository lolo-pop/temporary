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
import csv

def evaluate_model(evaluate_config):
    filePath = evaluate_config["data_path"]
    completeInvocationTrace = [] # 完整的每个函数调用轨迹
    functionID = [] # function hash value
    functionStartTime = []
    with open(filePath, 'r', newline='') as f:
        reader = csv.reader(f)
        for row in reader:
            invocationList = []
            for i in range(len(row)):
                if i == 0:
                    functionID.append(row[i])
                else: 
                    invocationList.append(int(row[i]))
            completeInvocationTrace.append(invocationList)
            functionStartTime.append(pd.to_datetime(0, unit = "s"))
    prediction_length = evaluate_config["prediction_length"]
    model_path = evaluate_config["model_path"] 
    result_path = result_path = evaluate_config["evaluate_result_path"] 
    point_result_path = evaluate_config["point_result_path"] 
    os.makedirs(result_path, exist_ok=True)
    os.makedirs(point_result_path, exist_ok=True)   
    for model in config.evaluate_model:
        global_model_result = {
            link: {
                "point_forecast_result": [],
                "observation": []
            }
            for link in functionID
        }
        evaluation_result = {
            "agg_metrics": [],
            "item_metrics": []
        }
    
        for i, invocationList in enumerate(completeInvocationTrace):
            test_pairs = []

            test_data_length = int(evaluate_config["train_ds_ratio"] * len(invocationList)) + prediction_length

            while test_data_length <= len(invocationList):
                test_pairs.append((functionStartTime[i], invocationList[:test_data_length]))
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
            funcid = functionID[i]

            # 记录点预测结果，作为global model的输出
            for j, probabilistic_f in enumerate(forecasts):
                global_model_result[funcid]["point_forecast_result"].append(probabilistic_f.quantile(0.5).tolist())
                global_model_result[funcid]["observation"].append(
                    np.array(tss[j][-prediction_length:]).reshape(-1).tolist())

            # evaluate概率预测的结果
            # num_workers默认为CPU数，测试设备为一个16核的cpu
            evaluator = Evaluator(quantiles=config.evaluation_quantile, num_workers=16)
            agg_metrics, item_metrics = evaluator(tss, forecasts)

            evaluation_result["agg_metrics"].append(agg_metrics)
            evaluation_result["item_metrics"].append(item_metrics.to_json(orient='records'))

        with open(result_path / f"{model}.json", "w") as f:
            f.write(json.dumps(evaluation_result, indent=4, sort_keys=True))
        with open(point_result_path / f"{model}.json", "w") as f:
            f.write(json.dumps(global_model_result, indent=4, sort_keys=True))

    
if __name__ == "__main__":
    evaluate_model(config.dataset_meta_data)
