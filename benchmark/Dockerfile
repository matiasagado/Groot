# Use a Python base image
FROM python:3.11-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the necessary files into the container

COPY src/multiple-testcase.py /app
COPY data/normal_logs.txt /app
COPY data/error_logs.txt /app

# Install any necessary dependencies (if you have any)
# RUN pip install some-package

# Specify the command to run when the container starts
CMD python3 multiple-testcase.py --num_test_cases 5 --lower_bound 5 --upper_bound 20
