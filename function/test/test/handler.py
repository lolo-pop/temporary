
import json
import time
def handle(event, context):
    print(type(event.body))
    data = event.body.decode()
    json_data = json.loads(data)
    for image in json_data:
        image["data"] = "processed"
    time.sleep(2)
    return {
        "statusCode": 200,
        "body": json.dumps(json_data)
    }
