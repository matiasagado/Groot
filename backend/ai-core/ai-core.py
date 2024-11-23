import os
import redis
import time
import re
import requests
import logging
import uuid
import json
from typing import Optional, Dict, Any
from requests.exceptions import RequestException
from redis.exceptions import RedisError
from tenacity import retry, stop_after_attempt, wait_exponential
from clickhouse_driver import Client


logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s : %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

logger = logging.getLogger(__name__)

REDIS_CONFIG = {
    "host": os.getenv("REDIS_HOST", "localhost"),
    "port": int(os.getenv("REDIS_PORT", "6379")),
    "db": int(os.getenv("REDIS_DB", "0")),
    "decode_responses": True,
}

CLICKHOUSE_CONFIG = {
    "host": os.getenv("CLICKHOUSE_HOST", "poc-clickhouse"),
    "port": int(os.getenv("CLICKHOUSE_PORT", "9000")),
    "database": os.getenv("CLICKHOUSE_DB", "default"),
    "user": os.getenv("CLICKHOUSE_USER", "default"),
    "password": os.getenv("CLICKHOUSE_PASSWORD", "password"),
}

OAI_API_URL = os.getenv('OAI_API_URL')
OAI_TOKEN = os.getenv('OAI_TOKEN')
ONE_SHOT_PROMPT = """Please classify if the INPUT log line is an error, classifying it as INFO or ERROR. Please end the response in this format `CLASSIFICATION: INFO` or `CLASSIFICATION: ERROR`.
INPUT: 
```
{input}
```
"""

def extract_classification(result_text: str) -> Optional[str]:
    """Extract classification from API response."""
    try:
        classification = re.search(r"CLASSIFICATION: (\w+)", result_text)
        return classification.group(1) if classification else None
    except (AttributeError, IndexError) as e:
        logger.error(f"Error extracting classification: {e}")
        return None

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10),
    reraise=True
)
def classify_log_line(log_line: str, prompt_template: str) -> Optional[Dict[str, Any]]:
    """Classify a log line with retry logic."""
    if not all([OAI_API_URL, OAI_TOKEN]):
        raise ValueError("Missing required environment variables: OAI_API_URL or OAI_TOKEN")

    prompt = prompt_template.format(input=log_line)
    payload = {"max_tokens": 50, "messages": [{"role": "user", "content": prompt}]}
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {OAI_TOKEN}",
    }

    try:
        start_time = time.time()
        res = requests.post(OAI_API_URL, json=payload, headers=headers, timeout=10)
        execution_time = time.time() - start_time
        
        res.raise_for_status()
        prompt_result = res.json()["choices"][0]["message"]["content"]
        classification = extract_classification(prompt_result)
        
        if classification is None:
            logger.error(f"Classification not found in response: {prompt_result}")
            return None

        return {
            "log_line": log_line,
            "prompt_template": prompt_template,
            "result": prompt_result,
            "classification": classification,
            "execution_time": execution_time,
        }

    except RequestException as e:
        logger.error(f"Request failed: {e}")
        raise  # Let retry handle it

def send_to_clickhouse(classified_log):
    """Directly update the log entry in ClickHouse with classification result."""
    # Connect to ClickHouse
    client = Client(**CLICKHOUSE_CONFIG)

    # extracted log line to get the UUID
    log_line = classified_log["log_line"]

    try:
        log = json.loads(log_line)
    except json.JSONDecodeError as e:
        logger.error(f"Failed to decode log_line as JSON: {e}")
        return
    
    uuid_value = log.get("uuid")
    logger.info(f"UUID type: {type(uuid_value)}, UUID value: {uuid_value}")
    if uuid_value is None:
        logger.error("UUID not found in log_line")
        return

    classification = classified_log["classification"]
    
    logger.info(f"Updating ai_classified_level for UUID {uuid_value} to {classification}")

    query = """
    ALTER TABLE user_log_data
    UPDATE ai_classified_level = '{classification}'
    WHERE uuid = '{uuid_value}'
    """
    query = query.format(classification=classification, uuid_value=uuid_value)

    try:
        client.execute(query)
        logger.info(f"Successfully updated log for UUID {uuid_value} in ClickHouse.")
    except Exception as e:
        logger.error(f"Failed to update ClickHouse for UUID {uuid_value}: {e}")

class RedisQueue:
    """Redis queue handler with connection management."""
    def __init__(self, queue_name: str = "log_queue"):
        self.queue_name = queue_name
        self.redis_client = None
        self._connect()

    def _connect(self) -> None:
        """Establish Redis connection with error handling."""
        try:
            if not self.redis_client:
                self.redis_client = redis.Redis(**REDIS_CONFIG)
                # Test the connection
                self.redis_client.ping()
                logger.info("Successfully connected to Redis")
        except RedisError as e:
            logger.error(f"Failed to connect to Redis: {e}")
            raise

    def _ensure_connection(self) -> None:
        """Ensure Redis connection is active."""
        try:
            if not self.redis_client or not self.redis_client.ping():
                self._connect()
        except RedisError as e:
            logger.error(f"Redis connection check failed: {e}")
            self._connect()

    def fetch_items(self):
        """Generator to fetch items from Redis queue with connection handling."""
        while True:
            try:
                self._ensure_connection()
                # Block until an item is available (timeout=0 means block indefinitely)
                result = self.redis_client.blpop(self.queue_name, timeout=0)
                if result:
                    yield result[1]
            except RedisError as e:
                logger.error(f"Error fetching from Redis queue: {e}")
                # Wait before retry to avoid tight loop
                time.sleep(5)
                continue
            except Exception as e:
                logger.error(f"Unexpected error in fetch_items: {e}")
                time.sleep(5)
                continue

def main():
    """Main application loop with error handling."""
    logger.info("STARTING AI CORE")
    
    try:
        queue = RedisQueue()
        
        for item in queue.fetch_items():
            try:
                logger.info(f"Processing item: {item}")
                result = classify_log_line(item, ONE_SHOT_PROMPT)
                
                if result:
                    logger.info(f'{result["classification"]}: {result["log_line"]}')
                    
                    send_to_clickhouse(result)  # Send as a list of logs
                else:
                    logger.warning(f"Failed to classify item: {item}")
                
            except Exception as e:
                logger.error(f"Error processing item {item}: {e}")
                continue  # Continue with next item

    except KeyboardInterrupt:
        logger.info("Shutting down gracefully...")
    except Exception as e:
        logger.error(f"Critical error in main loop: {e}")
        raise

if __name__ == "__main__":
    main()
