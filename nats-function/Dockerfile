FROM python:3.9-slim-buster

WORKDIR /app

COPY . .

RUN pip install nats-py

CMD ["python", "nats-function/handler.py"]