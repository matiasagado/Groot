import json
import os
import requests
import time
import logging
import re

# Setup logger
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Example prompt template for OpenAI compatible API
OAI_API_URL = "http://bugatti.internal-headscale.ucaia.com:8920/v1/chat/completions"
OAI_TOKEN = "aa"
ONE_SHOT_PROMPT = """Please classify if the INPUT log line is an error, classifying it as INFO or ERROR. Please end the response in this format CLASSIFICATION: INFO or CLASSIFICATION: ERROR.
INPUT: 
{input}

"""

def extract_classification(result_text):
    classification = re.search(r"CLASSIFICATION: (\w+)", result_text)
    return classification.group(1) if classification else None

def classify_log_line(log_line, prompt_template):
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
            "classification": classification,
        }

    except requests.RequestException as e:
        logger.error(f"Request failed: {e}")
        return None

def save_classification_to_json(results, output_file):
    try:
        with open(output_file, 'w') as json_file:
            json.dump(results, json_file, indent=4)  # Save all results to a new file
            logger.info(f"Saved classification results to {output_file}")
    except Exception as e:
        logger.error(f"Error saving to JSON file: {e}")

def process_test_case_file(file_path, output_directory):

    try:
        with open(file_path, 'r') as file:
            test_case_data = json.load(file)
            results = []

            for entry in test_case_data:
                log_line = entry.get('log', '')
                if log_line:  # Ensure we're not processing empty logs
                    logger.info(f"Classifying log line: {log_line}")
                    result = classify_log_line(log_line, ONE_SHOT_PROMPT)
                    if result:
                        results.append(result)

            # Generate output file name based on input file name
            file_name = os.path.basename(file_path)
            output_file = os.path.join(output_directory, f"{file_name}_classified.json")
            save_classification_to_json(results, output_file)

    except FileNotFoundError:
        logger.error(f"File not found: {file_path}")
    except Exception as e:
        logger.error(f"Error processing test case file: {e}")

def process_all_test_cases(directory_path, output_directory):
    try:
        if not os.path.exists(output_directory):
            os.makedirs(output_directory)  # Create the output directory if it doesn't exist

        # Iterate through all files in the test case directory
        for file_name in os.listdir(directory_path):
            if file_name.endswith(".json"):
                file_path = os.path.join(directory_path, file_name)
                logger.info(f"Processing test case file: {file_path}")
                process_test_case_file(file_path, output_directory)
            break

    except Exception as e:
        logger.error(f"Error processing test case directory: {e}")

if __name__ == "__main__":
    # Directory containing the test case files
    test_case_directory = "test_cases_json"
    # Directory where the output classified files will be saved
    output_directory = "classified_test_cases"

    process_all_test_cases(test_case_directory, output_directory)
