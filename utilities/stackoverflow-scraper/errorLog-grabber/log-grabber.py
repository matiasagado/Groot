import os
import redis
import re

# Connect to Redis. Defaults match the scraper's own docker-compose, which
# exposes redis on localhost:6379. Override via env when running against a
# different host.
r = redis.Redis(
    host=os.getenv("REDIS_HOST", "localhost"),
    port=int(os.getenv("REDIS_PORT", "6379")),
)

# Ping the server to ensure the connection works
try:
    r.ping()
    print("Connected to Redis server!")
except redis.ConnectionError:
    print("Failed to connect to Redis server.")

hash_key = "so_data"  # Replace with your hash key
#field = 74613757  # Replace with the specific field you're looking for
field = 23262663
hash_value = r.hget(hash_key, field)

def extract_error_logs(text):
    # Define the regex pattern for extracting Nginx error log lines
    error_log_pattern = r'nginx:\s+[^\n]+'
    
    # Use the findall method to extract all matching error log lines
    error_logs = re.findall(error_log_pattern, text)
    
    return error_logs

decoded_value = hash_value.decode('utf-8')
print(decoded_value)
extracted_logs = extract_error_logs(decoded_value)

# Print the extracted error log lines
for log in extracted_logs:
    print(log)



