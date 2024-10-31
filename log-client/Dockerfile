FROM python:3.11-alpine

COPY src/requirements.txt /app/requirements.txt

WORKDIR /app

RUN pip3 install -r ./requirements.txt

COPY src /app

RUN python3 /app/logs-bootstrapper.py
