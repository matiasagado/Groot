import json
import os
import requests
import time
import logging
import re

# Setup logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Example prompt template for Tabby API
TABBY_API_URL = os.getenv('BERT_API_URL')
TABBY_API_KEY = os.getenv('BERT_API_KEY')
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
        "Authorization": f"Bearer {TABBY_API_KEY}",
    }

    try:
        start_time = time.time()
        res = requests.post(TABBY_API_URL, json=payload, headers=headers, timeout=10)
        execution_time = time.time() - start_time
        res.raise_for_status()

        prompt_result = res.json()["choices"][0]["message"]["content"]
        classification = extract_classification(prompt_result)

        if classification is None:
            logger.error(f"Classification not found in response: {prompt_result}")
            return None

        return {
            "log_line": log_line,
            "classification": classification,
        }

    except requests.RequestException as e:
        logger.error(f"Request failed: {e}")
        return None

def save_classification_to_json(result):
    # Get the root directory and specify the path to the file in the root folder
    root_path = os.path.dirname(os.path.abspath(__file__))
    json_file_path = os.path.join(root_path, 'classification_results.json')
    
    try:
        # Check if file exists and load existing content
        if os.path.exists(json_file_path):
            with open(json_file_path, 'r') as json_file:
                try:
                    existing_data = json.load(json_file)
                    # Ensure that the existing data is a list
                    if not isinstance(existing_data, list):
                        existing_data = []
                except json.JSONDecodeError:
                    # If the file exists but is invalid, treat it as empty
                    existing_data = []
        else:
            existing_data = []

        # Append the new result to the existing data (which is ensured to be a list)
        existing_data.append(result)

        # Write the updated data to the file in pretty format (new lines, indentations)
        with open(json_file_path, 'w') as json_file:
            json.dump(existing_data, json_file, indent=4)  # Use indent=4 for pretty formatting
            logger.info(f"Saved classification result: {result}")
    except Exception as e:
        logger.error(f"Error saving to JSON file: {e}")

def clear_json_file(file_path):
    with open(file_path, 'w') as f:
        f.write('')  # Writing an empty string clears the file

def process_log_file(file_path):
    try:
        with open(file_path, 'r') as file:
            for line in file:
                line = line.strip()
                if line:  # Ensure we're not processing empty lines
                    logger.info(f"Classifying log line: {line}")
                    result = classify_log_line(line, ONE_SHOT_PROMPT)
                    if result:
                        save_classification_to_json(result)

    except FileNotFoundError:
        logger.error(f"File not found: {file_path}")
    except Exception as e:
        logger.error(f"Error processing log file: {e}")

if __name__ == "__main__":

    log_file_path = "output_logs.txt"
    output_file_path = "classification_results.json"
    clear_json_file(output_file_path)
    process_log_file(log_file_path)