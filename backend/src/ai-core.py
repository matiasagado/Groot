import os
import redis
import time
import re
import requests
import logging

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s : %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

# Example usage
logger = logging.getLogger(__name__)

TABBY_CHAT_COMPLETION_URL = (
    f"http://{os.getenv('TABBY_HOST')}:{os.getenv('TABBY_PORT')}/v1/chat/completions"
)
ONE_SHOT_PROMPT = """Please classify if the INPUT log line is an error, classifying it as INFO or ERROR. Please end the response in this format `CLASSIFICATION: INFO` or `CLASSIFICATION: ERROR`.
INPUT: 
```
{input}
```
"""


def extract_classification(result_text):
    classification = re.search(r"CLASSIFICATION: (\w+)", result_text)
    return classification.group(1) if classification else None


def classify_log_line(log_line, prompt_template):
    prompt = prompt_template.format(input=log_line)
    payload = {"max_tokens": 50, "messages": [{"role": "user", "content": prompt}]}
    headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {os.getenv('TABBY_API_KEY')}",
    }

    try:
        start_time = time.time()
        res = requests.post(
            TABBY_CHAT_COMPLETION_URL, json=payload, headers=headers, timeout=10
        )
        execution_time = time.time() - start_time
        res.raise_for_status()
        prompt_result = res.json()["choices"][0]["message"]["content"]
        classification = extract_classification(prompt_result)
        if classification is None:
            logger.error(f"Classification not found in response: {prompt_result}")
            # TODO: Retry since there was a failure

    except requests.RequestException as e:
        logger.error(f"Request failed: {e}")
        # TODO: Decide on retry, skip, or exit strategy
        return None

    return {
        "log_line": log_line,
        "prompt_template": prompt_template,
        "result": prompt_result,
        "classification": classification,
        "execution_time": execution_time,
    }


def fetch_items_from_redis_queue():
    # Adjusted for environment variables and added try-except
    try:
        r = redis.Redis(
            host=os.getenv("REDIS_HOST"),
            port=os.getenv("REDIS_PORT"),
            decode_responses=True,
        )
        while True:
            result = r.blpop("log_queue", 0)
            if result:
                yield result[1]
    except redis.RedisError as e:
        logger.error(f"Redis error: {e}")
    except:
        logger.error(f"Unexpected error: {e}")
        # TODO: Handle error or retry logic

def main():
    logger.info("STARTING AI CORE")

    for item in fetch_items_from_redis_queue():
        logger.info("checking...")
        logger.info(f"Processing item: {item}")
        result = classify_log_line(item, ONE_SHOT_PROMPT)
        logger.info(f'{result["classification"]}: {result["log_line"]}')


if __name__ == "__main__":
    main()
