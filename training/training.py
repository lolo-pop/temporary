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
import csv
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


def train_model(train_config):
  filePath = train_config["data_path"]
  completeInvocationTrace = []
  with open(filePath, 'r', newline='') as f:
    reader = csv.reader(f)
    for row in reader:
      invocationList = []
      for i in range(len(row)):
        if i == 0:
          continue
        invocationList.append(float(row[i]))
      # print(invocationList)
      completeInvocationTrace.append(invocationList)
  train_ds = ListDataset(
    data_iter = [
      {
        FieldName.TARGET: completeInvocationTrace[i][:int(train_config["train_ds_ratio"]*len(completeInvocationTrace[i]))], 
        FieldName.START: pd.to_datetime(0, unit="s")
      }
      for i in range(len(completeInvocationTrace))
    ],
    freq = train_config["frequency"]
  )
  prediction_length = train_config["prediction_length"]
  Estimators = {
    "DeepAR": DeepAREstimator(
            prediction_length=prediction_length,
            freq=train_config["frequency"],
            context_length = 10
        ),
    "Transformer": TransformerEstimator(
            prediction_length=prediction_length,
            freq=train_config["frequency"],
            context_length = 10
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
    model_output_path = train_config["model_path"] / f"{model_name}"
    predictor = Estimator.train(train_ds)
    os.makedirs(model_output_path, exist_ok=True)
    predictor.serialize(model_output_path)

  # 使用多线程并行训练
  # for model_name in config.train_model:
  pool = ThreadPoolExecutor(max_workers=32)
  pool.map(train_process, config.train_model)
  
if __name__ == "__main__":
  train_model(config.dataset_meta_data)
        
    
    
    
    