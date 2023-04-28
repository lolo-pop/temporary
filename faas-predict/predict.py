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
    sequence = sequence.split(',')  
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
      'start_date' : str(forecasts[0].start_date),
      '0.1': json.dumps(float(forecasts[0].quantile(0.1).mean())), 
      '0.2': json.dumps(float(forecasts[0].quantile(0.2).mean())),
      '0.3': json.dumps(float(forecasts[0].quantile(0.3).mean())), 
      '0.4': json.dumps(float(forecasts[0].quantile(0.4).mean())),
      '0.5': json.dumps(float(forecasts[0].quantile(0.5).mean())), 
      '0.6': json.dumps(float(forecasts[0].quantile(0.6).mean())),
      '0.7': json.dumps(float(forecasts[0].quantile(0.7).mean())), 
      '0.8': json.dumps(float(forecasts[0].quantile(0.8).mean())),
      '0.9': json.dumps(float(forecasts[0].quantile(0.9).mean())),
      'mean': json.dumps(float(np.mean(forecasts[0].mean)))
    }
    print(result_dict)
    json_str = json.dumps(result_dict)

    # Return the predicted value in the response
    return jsonify({'predicted_value': json_str})

if __name__ == '__main__':
    app.run(debug=True)
    
    