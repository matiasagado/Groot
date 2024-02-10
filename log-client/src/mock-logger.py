import time
import os
import redis
import argparse


# Get Redis host and port from environment variables
redis_host = os.getenv("REDIS_HOST")
redis_port = os.getenv("REDIS_PORT")

# Parse command line arguments
parser = argparse.ArgumentParser()
parser.add_argument("file", help="Path to the file to read")
args = parser.parse_args()

# Connect to Redis
r = redis.Redis(host=redis_host, port=redis_port)

# Read the file line by line
with open(args.file, "r") as file:
    for line in file:
        # Put each line into the Redis queue
        r.lpush("log_queue", line.strip())

print("Items added to Redis queue successfully!", flush=True)

time.sleep(15)
