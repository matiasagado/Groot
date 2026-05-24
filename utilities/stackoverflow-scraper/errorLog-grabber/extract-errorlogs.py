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

list_key = "[nginx]"  # Replace with your list key

# Change the range here for looking at different indexes. (E.G: looking at 0 index to 5)
list_elements = r.lrange(list_key, 0, 100)

# Decode byte values (if they are stored as bytes)
decoded_list_elements = [element.decode('utf-8') for element in list_elements]

# Print the list elements
print("List elements:")

for element in decoded_list_elements:
    print(element)

def extract_error_logs(text):
    # Define the regex pattern for extracting Nginx error log lines
    error_log_pattern = r'nginx:\s+[^\n]+'
    #error_log_pattern = r'^\[\s*\d{1,}\.\d{6}\]\s.*$'
    
    # Use the findall method to extract all matching error log lines
    error_logs = re.findall(error_log_pattern, text)
    
    return error_logs

# Initialize an empty list to hold all extracted logs
extracted_logs = []

# Hash key that contains stack overflow posts in Redis (replace with your actual hash key)
hash_key = "so_data"
for field in decoded_list_elements:
    # Retrieve the value from the hash for the specific field
    hash_value = r.hget(hash_key, field)
    # Check if the value exists in the hash
    if hash_value:
        # Decode the hash value from bytes to string
        decoded_value = hash_value.decode('utf-8')
        
        # Extract error logs from the decoded value
        logs = extract_error_logs(decoded_value)
        
        # Append the extracted logs to the main list
        extracted_logs.extend(logs)
    else:
        print(f"Field '{field}' not found in hash '{hash_key}'")

# Print the extracted error log lines
print("Extracted error log lines:")
for log in extracted_logs:
    print(log)
    print("\n\n")



