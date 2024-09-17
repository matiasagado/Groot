FROM python:3.10

WORKDIR /app

COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

COPY . /app/

CMD ["python", "read-and-get-data.py"]  
# CMD ["python", "read-and-get-data.py", $REDIS_CONTAINER_NAME, $REDIS_SERVICE_PORT, $QUESTION_QUEUE_NAME, $SO_DB_NAME]  