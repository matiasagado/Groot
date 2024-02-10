import os, json

from flask import Flask, request
import redis

app = Flask(__name__)

# Get Redis host and port from environment variables
redis_host = os.getenv("REDIS_HOST")
redis_port = os.getenv("REDIS_PORT")

# Connect to Redis
r = redis.Redis(host=redis_host, port=redis_port)

# This is used to view the incoming logs from Vector
@app.route('/test/<path:variable_part>', methods=['POST'])
def test(variable_part):
    variable_part = variable_part.replace('/', '_')
    combined_string = (
        f"Variable part of the URL: {variable_part}\n"
        f"request.headers: {request.headers}\n"
        f"request.method: {request.method}\n"
        # f"request.json: {request.json}"
    )
    print(combined_string)

    with open(f'./test-{variable_part}.json', 'w+') as f:
        f.write(json.dumps(request.json, indent=4))

    print(f'-'*50)

    return 'Test', 200

@app.route('/logs/ingest', methods=['POST'])
def logs_ingest():
    # Parse JSON body of the request
    vector_logs = request.json
    
    for log in vector_logs:
        if log.keys().contains('nginx'):
            r.lpush('nginx_logs', log['original_message'])
            # TODO: push along the rest of the information
            # to an ingesting service that will export to the ClickHouse DB

    return 'Received', 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=9500)
    
