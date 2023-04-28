FROM python:3.7.16

RUN pip install flask \
                gluonts \
                pandas \
                "gluonts[mxnet]"

WORKDIR /home/app 
COPY . .
CMD ["python", "predict.py"]