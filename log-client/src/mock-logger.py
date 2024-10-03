import time
import os
import redis
import argparse
from ingestor import read_log_file


# Get Redis host and port from environment variables
redis_host = os.getenv("REDIS_HOST")
redis_port = os.getenv("REDIS_PORT")

# Parse command line arguments
parser = argparse.ArgumentParser()
parser.add_argument("file", help="Path to the file to read")
args = parser.parse_args()

# Connect to Redis
r = redis.Redis(host=redis_host, port=redis_port)

# testing if it would read the test.txt file
# TODO: Improve read_log_file
# read_log_file("/app/test.txt")
# r.lpush("log_queue", "test")

# Read the file line by line from 5-line.log inside logs/dmesg
with open(args.file, "r") as file:
    for line in file:
        # Put each line into the Redis queue
        print("Adding items: ", line, flush=True)
        r.lpush("log_queue", line.strip())

# This would read the test.txt and add it into the redis queue
# with open("/app/test.txt", "r") as file:
#     for line in file:
#         # Put each line into the Redis queue
#         r.lpush("log_queue", line.strip())

print("Items added to Redis queue successfully!", flush=True)

time.sleep(15)
