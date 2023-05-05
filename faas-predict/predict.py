from flask import Flask, request, jsonify
from gluonts.model.predictor import Predictor
from gluonts.dataset.field_names import FieldName
import numpy as np
from pathlib import Path
import pandas as pd
from gluonts.dataset.common import ListDataset
import json

from gluonts.model.forecast import QuantileForecast

# Load the pre-trained model
predictor = Predictor.deserialize(Path("TFT"))

app = Flask(__name__)

@app.route('/predict', methods=['POST'])
def predict():
    # Get the monitoring sequence from the request
    data = request.json
    sequence = data['monitoring_sequence']
    # sequence = sequence.split(',')  
    monitoring_sequence = np.array([float(n) for n in sequence])
    
    # Convert the monitoring sequence to GluonTS format
    monitoring_data = ListDataset (data_iter =
                                   [
                                     {FieldName.START: pd.to_datetime(0, unit="s"), FieldName.TARGET: monitoring_sequence}
                                     ],  
                                   freq = "30s"
                                   )

    # Make predictions for the next time step
    forecast_it = predictor.predict(dataset=monitoring_data, num_samples=500)

    # Get the predicted value for the next time step
    forecasts = list(forecast_it)
    # predicted_value = forecasts[0].mean(axis=0)[-1]
    print(np.mean(forecasts[0].mean))
    result_dict = {
      'function_name': data['function_name'],
      'start_date' : str(forecasts[0].start_date),
      'quantile0.1': float(forecasts[0].quantile(0.1).mean()), 
      'quantile0.2': float(forecasts[0].quantile(0.2).mean()),
      'quantile0.3': float(forecasts[0].quantile(0.3).mean()), 
      'quantile0.4': float(forecasts[0].quantile(0.4).mean()),
      'quantile0.5': float(forecasts[0].quantile(0.5).mean()), 
      'quantile0.6': float(forecasts[0].quantile(0.6).mean()),
      'quantile0.7': float(forecasts[0].quantile(0.7).mean()), 
      'quantile0.8': float(forecasts[0].quantile(0.8).mean()),
      'quantile0.9': float(forecasts[0].quantile(0.9).mean()),
      'mean': float(np.mean(forecasts[0].mean))
    }
    print(result_dict)
    json_str = json.dumps(result_dict, ensure_ascii=False)

    # Return the predicted value in the response
    return json_str

if __name__ == '__main__':
    app.run(debug=True)
    
    