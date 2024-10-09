import redis
import re

# Connect to Redis
r = redis.Redis(host='logparse-nix.internal-headscale.ucaia.com', port=6379)

# Ping the server to ensure the connection works
try:
    r.ping()
    print("Connected to Redis server!")
except redis.ConnectionError:
    print("Failed to connect to Redis server.")

hash_key = "so_data"  # Replace with your hash key
field = 74613757  # Replace with the specific field you're looking for
hash_value = r.hget(hash_key, field)

def extract_error_logs(text):
    # Define the regex pattern for extracting Nginx error log lines
    error_log_pattern = r'nginx:\s+[^\n]+'
    
    # Use the findall method to extract all matching error log lines
    error_logs = re.findall(error_log_pattern, text)
    
    return error_logs

decoded_value = hash_value.decode('utf-8')
#print(decoded_value)
extracted_logs = extract_error_logs(decoded_value)

# Print the extracted error log lines
for log in extracted_logs:
    print(log)



