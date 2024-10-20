## **Overview**

'benchmark.py' automates the generation of log test cases, classifies them using an AI model, and evaluates the model’s performance. The workflow consists of three main steps:

1. **Test Case Generation**: Randomly selects logs from data/normal_logs.txt and data/error_logs.txt to create test cases, which are saved in data/test_cases_json.
2. **Classification**: The AI model classifies the log entries. Results are saved in data/classified_test_cases.
3. **Scoring**: Compares the AI's classification to the ground truth, calculating metrics such as accuracy, precision, recall, and F1 score. Results are saved in data/src/metric_results.json.

## **Output Files**
**Test Cases**: data/test_cases_json.
**Classified Results**: data/classified_test_cases
**Metrics**: data/src/metric_results.json