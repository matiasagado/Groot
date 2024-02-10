FROM python:3.11-alpine

COPY src/requirements.txt /app/requirements.txt

WORKDIR /app

RUN pip3 install -r ./requirements.txt

COPY src /app

CMD python3 /app/ai-core.py