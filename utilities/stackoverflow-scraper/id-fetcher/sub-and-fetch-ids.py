import traceback
import redis
import json
import time
import re
import os
import sys
from pystackapi import Site
from pystackapi.errors import HttpError
from pystackapi.sites import StackOverflow



r = redis.Redis(host=os.environ['REDIS_CONTAINER_NAME'], port=os.environ['REDIS_SERVICE_PORT'], db=0, decode_responses=True)

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


def get_or_create_list_length(key):
    # Check if the list exists
    if not r.exists(key):
        # If the list doesn't exist, create it by pushing an item and then removing it
        r.rpush(key, "")  # Push a placeholder item to create the list
        r.lpop(key)         # Remove the placeholder item to ensure the list is empty but exists
    
    # Return the length of the list
    return r.llen(key)


def append_to_list(key, value):
    # Append a string to the end of the list
    r.rpush(key, value)


def fetch_all_advanced(site, query):
    pagesize = 100
    # max_pages = 7 # hardcoded limit for now
    max_pages = sys.maxsize
    ## I am using the raw query as a list index. 
    length_list = get_or_create_list_length(query)

    print(length_list)

    if length_list >= pagesize*max_pages:
        print("Already Gathered Max")
    else:
        ## TODO check this works with all edge cases, I think first page is 
        page = int(length_list/pagesize + 1)

        print(page)


        try:
            while True:
                print("query %s page %s pagesize %s" % (query, page, pagesize))
                search_results = site.advanced_search(q=query, page=page, pagesize=pagesize, key=os.environ['SO_APP_KEY'])
                num_results = len(search_results)

                print(num_results)
                
                for item in search_results: 
                    # item_obj = json.loads(item)

                    if (item["question_id"] is not None):
                        append_to_list(query, item["question_id"])
                        append_to_list(os.environ['QUESTION_QUEUE_NAME'], item["question_id"])
                    
                
                # If less than 100 items are returned, it means we've reached the last page
                if num_results < pagesize or page >= max_pages:
                    # TODO add a marker to the end of the list to mark as complete
                    break
                
                
                page += 1
                
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



def handle_message(message):
    print(f"Received: {message}")

    # msg_obj = json.loads(message)
    
    site = Site(StackOverflow)

    print(message)

    if (message["type"] == "message"):
        try:
            fetch_all_advanced(site, message["data"])   
        except Exception as e:
            print(f"Fetch All Advanced failed Error: {e}")
            print(traceback.format_exc())
            
    
def subscribe_to_channel():
    while True:
        try:
            # r = redis.Redis(host='redis', port=6379, db=0)
            p = r.pubsub()
            # p.subscribe(**{'query-input': handle_message})
            p.subscribe(os.environ['INPUT_QUERY_CHANNEL_NAME']) #TODO change to env
            print("Subscribed to %s. Waiting for messages..." % os.environ['INPUT_QUERY_CHANNEL_NAME'])

            ## TODO the issue is currently the subscribe event is not getting translated to the query right. 
            ## !!!  --- now its a typing issue somewhere after handle_message
            for message in p.listen():
                handle_message(message) 

            # for message in p.listen():
            #     print("caught")
            #     print(message) 
            # while True:
            #     message = p.get_message()
            #     if message:
            #         handle_message(message)
            #     time.sleep(1)
        except Exception as e:
            print(f"Subscription failed, retrying in 5 seconds. Error: {e}")
            time.sleep(5)


if __name__ == "__main__":

    print("ID Fetcher Starting...")
    time.sleep(5) ## TODO remove this if possible. 
    subscribe_to_channel()  