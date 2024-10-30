# Use a Python base image
FROM python:3.11-alpine

COPY src/requirements.txt /app/requirements.txt

WORKDIR /app

RUN pip3 install -r ./requirements.txt

# Copy the necessary files into the container

COPY src/multiple_testcase.py /app
COPY src/benchmark.py /app
COPY src/scoring_metric.py /app
COPY src/ai_core.py /app
COPY data/normal_logs.txt /app
COPY data/error_logs.txt /app

# Install any necessary dependencies (if you have any)
# RUN pip install some-package

# Specify the command to run when the container starts
CMD python3 benchmark.py
