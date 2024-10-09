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
field = 35538757  # Replace with the specific field you're looking for
hash_value = r.hget(hash_key, field)

if hash_value:
    print(f"Value for field '{field}' in hash '{hash_key}': {hash_value}")
else:
    print(f"No value found for field '{field}' in hash '{hash_key}'")

