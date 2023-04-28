from gluonts.model.predictor import Predictor
from gluonts.dataset.field_names import FieldName
import numpy as np
from pathlib import Path
import pandas as pd
import sys
from gluonts.dataset.common import ListDataset
# Load the pre-trained model
predictor = Predictor.deserialize(Path("/home/rongch05/openfaas/training/predict_models_by_appID/TFT"))
args = sys.argv[1]
print(args)
args = args.split(",")
trace = []
for n in args:
  trace.append(int(n))
# Define the monitoring sequence
monitoring_sequence = np.array(trace)
# 140,503,411,388,320,295,288,150,203,73,735,577,436,378,469,307,110,95,140
# Convert the monitoring sequence to GluonTS format
print(pd.to_datetime(0, unit="s"))
monitoring_data = ListDataset (data_iter =
                               [
                                 {FieldName.START: pd.to_datetime(0, unit="s"), FieldName.TARGET: monitoring_sequence}
                                 ],  
                               freq = "30s"
                               )

# Make predictions for the next time step
forecast_it = predictor.predict(dataset=monitoring_data, num_samples=500)

# Print the predicted value for the next time step
forecasts = list(forecast_it)
print(forecasts[0])
# print(f"Predicted value: {forecasts[0].samples.mean(axis=0)[-1]}")