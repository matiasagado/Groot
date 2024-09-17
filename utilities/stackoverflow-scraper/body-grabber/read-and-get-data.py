import traceback
import redis
import json
import time
import re
import os
from pystackapi import Site
from pystackapi.errors import HttpError
from pystackapi.sites import StackOverflow



def create_redis_connection(redis_hostname, redis_port, attempts=5, delay=5):
    """
    Attempt to create a Redis connection with retries.
    
    :param attempts: Number of connection attempts.
    :param delay: Delay between attempts in seconds.
    :return: Redis connection object if successful, None otherwise.
    """
    for attempt in range(attempts):
        try:
            return redis.Redis(host=redis_hostname, port=redis_port, db=0, decode_responses=True) # TODO move to env Var
        except redis.RedisError:
            if attempt < attempts - 1:
                print(f"Redis connection failed on attempt {attempt + 1}. Retrying in {delay} seconds...")
                time.sleep(delay)
            else:
                print("Failed to connect to Redis after several attempts.")
    return None

# r = redis.Redis(host='redis', port=6379, db=0, decode_responses=True)


# Function to sleep with periodic updates
def sleep_with_updates(wait_time, update_interval=30):
    start_time = time.time()
    while True:
        current_time = time.time()
        elapsed_time = current_time - start_time
        remaining_time = wait_time - elapsed_time
        
        if remaining_time > 0:
            print(f"Waiting... Time left: {int(remaining_time)} seconds")
            # Sleep for either the update interval or the remaining time, whichever is shorter
            time.sleep(min(update_interval, remaining_time))
        else:
            break

def fetch_question_data(site, qid, database_name):
    try:
        # Assuming 'key' should be 'qid' based on the context of your code
        exists = r.hexists(database_name, qid)

        if not exists:
            response = site.get_question(qid, filter="!T*hPNRA69ofM1izkPP", key=os.environ['SO_APP_KEY'])  # Filters to add bodies to response
            r.hset(database_name, qid, str(response))

    except HttpError as e:
        # Convert the error to a string to search within it
        e_str = str(e)
        print(e_str)  # Print the error message for debugging
        
        # Use regular expression to find the status code
        match = re.search(r'status code (\d+)', e_str)
        if match:
            status_code = int(match.group(1))  # Convert the matched status code to an integer
            print(f"Status Code: {status_code}")
            if "violation of backoff parameter" in e_str.lower() and status_code == 400:
                print("Violation of backoff parameter detected. Sleeping for 30 seconds.")
                sleep_with_updates(30)  # Sleep for 30 seco
                # Now you can handle the status code accordingly    
            elif status_code == 400:  # Checking for HTTP 400 error specifically

                # Requeue the ID for later processing
                # NOTE not pushing back if its not a API rate-limiting, in case its a bad formed request. 
                print(f"Re-queueing ID {qid} for later processing.")
                r.rpush(queue_name, qid)  # Push the ID back to the end of the list
        
                match = re.search(r'more requests available in (\d+) seconds', str(e))
                if match:
                    wait_time = int(match.group(1))
                    print(f"Rate limit exceeded. Waiting for {wait_time} seconds.")
                    sleep_with_updates(wait_time + 10)  # Adding a small buffer to ensure the limit is reset
                else:
                    print("HTTP Error 400 received, but unable to extract wait time.")
            else:
                raise  # Re-raise the exception if it's not a rate limit error

        else:
            print("Could not find status code in the error message.")


if __name__ == "__main__":

    # TODO change this to use hashes instead of lists? Something with unique keys so we dont have issues resuming?

    print("Body Grabber Starting...")
    time.sleep(5)

    r = create_redis_connection(os.environ['REDIS_CONTAINER_NAME'], os.environ['REDIS_SERVICE_PORT'])  # Use the new function to create a Redis connection
    if r is None:
        print("Exiting due to Redis connection failure.")
        exit(1)

    site = Site(StackOverflow) ## TODO Change to ENV Var
    queue_name = os.environ['QUESTION_QUEUE_NAME']
    database_name = os.environ['SO_DB_NAME']

    while True:
        # BLPOP key [key ...] timeout
        # Waits for the first element for an indefinite period (0 timeout)
        # Returns a tuple of (key, value) or None if the timeout occurs
        result = r.blpop(queue_name, timeout=0)
        
        if result:
            key, value = result
            print(f"Popped {value} from {key}")

            fetch_question_data(site, value, database_name)

        else:
            # This branch will not be reached because timeout=0 causes indefinite waiting
            print("Timeout occurred, but this should not happen.")