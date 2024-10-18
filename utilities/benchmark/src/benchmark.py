import os
from ai_core import process_all_test_cases  # From ai-core
from scoring_metric import load_classification_data, extract_labels, calculate_metrics, display_metrics, plot_confusion_matrix, save_results  # From scoring
from multiple_testcase import generate_test_cases, read_logs  # From the test case generator

def generate_test_cases_for_classification(normal_logs_file, error_logs_file, num_test_cases, lower_bound, upper_bound, test_case_output_dir):
    """
    Generate test cases using the logs provided, to be used in the classification.
    """
    print("Generating test cases...")
    
    # Read logs from files
    normal_logs = read_logs(normal_logs_file)
    error_logs = read_logs(error_logs_file)

    # Generate the test cases
    generate_test_cases(normal_logs, error_logs, num_test_cases, lower_bound, upper_bound, test_case_output_dir)

    print(f"{num_test_cases} test cases generated in {test_case_output_dir}")

def run_classification(test_case_dir, output_dir):
    """
    Run the ai-core.py classification process to classify logs and save the output.
    """
    print("Running classification process...")
    process_all_test_cases(test_case_dir, output_dir)
    print("Classification completed and results saved to:", output_dir)

def run_scoring(output_file):
    """
    Run the scoring metrics script using the classification results file.
    """
    # Load classification results from the output directory
    data = load_classification_data(output_file)

    if data:
        # Extract true and predicted labels
        true_labels, predicted_labels = extract_labels(data)

        # Calculate performance metrics
        accuracy, precision, recall, f1, conf_matrix = calculate_metrics(true_labels, predicted_labels)

        # Display the metrics and plot the confusion matrix
        display_metrics(accuracy, precision, recall, f1, conf_matrix)
        plot_confusion_matrix(conf_matrix)

        # Save the metrics results to a JSON file
        save_results(true_labels, predicted_labels, accuracy, precision, recall, f1, conf_matrix)
        print(f"Scoring results saved to 'metric_results.json'.")

def sync_process(normal_logs_file, error_logs_file, num_test_cases=5, lower_bound=5, upper_bound=20, test_case_dir="./utilities/benchmark/data/test_cases_json", output_dir="./utilities/benchmark/data/classified_test_cases"):
    """
    Sync the entire process:
    1. Generate test cases
    2. Run classification
    3. Run scoring and metrics evaluation
    """
    # Step 1: Generate test cases
    generate_test_cases_for_classification(normal_logs_file, error_logs_file, num_test_cases, lower_bound, upper_bound, test_case_dir)

    # Step 2: Run classification
    run_classification(test_case_dir, output_dir)

    # Step 3: Run scoring and metrics evaluation
    # output_file = os.path.join(output_dir, "test_classification_results.json")
    # run_scoring(output_file)

    # Step 3: Loop through each classified JSON file in the classified_test_cases directory
    print("Running scoring on classified test cases...")
    for file_name in os.listdir("./utilities/benchmark/data/classified_test_cases"):
        # Check if the file name matches the pattern "test_case_X.json_classified.json"
        if file_name.endswith("_classified.json") and file_name.startswith("test_case_"):
            classified_file_path = os.path.join("./utilities/benchmark/data/classified_test_cases", file_name)
            print(f"Processing file: {classified_file_path}")
            run_scoring(classified_file_path)  # Call your scoring function on each classified file

if __name__ == "__main__":
    # Define the paths to log files for test case generation
    normal_logs_file = os.path.join("./utilities/benchmark/data", 'normal_logs.txt')
    error_logs_file = os.path.join("./utilities/benchmark/data", 'error_logs.txt')
    
    # Start the process: generating test cases, running classification, and scoring
    sync_process(normal_logs_file, error_logs_file)
