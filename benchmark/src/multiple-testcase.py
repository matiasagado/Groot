import random
import os
import json
import argparse

def read_logs(file_path):
    """Reads logs from a file and returns them as a list of lines."""
    with open(file_path, 'r') as file:
        return file.readlines()

def sample_logs(logs, count):
    """Randomly selects a number of log lines from the list."""
    return random.sample(logs, count)

def create_log_entry(log_line, is_normal):
    """Creates a dictionary entry for a log line with a boolean flag."""
    return {
        "log": log_line.strip(),
        "is_normal": is_normal
    }

def create_log_output(normal_logs, error_logs, normal_count, error_count):
    """Creates a combined list of log entries with random order."""
    # Sample random log lines
    sampled_normal_logs = sample_logs(normal_logs, normal_count)
    sampled_error_logs = sample_logs(error_logs, error_count)
    
    # Create log entries
    log_entries = []
    for log in sampled_normal_logs:
        log_entries.append(create_log_entry(log, True))
    for log in sampled_error_logs:
        log_entries.append(create_log_entry(log, False))
    
    # Shuffle the log entries to randomize their order
    random.shuffle(log_entries)
    return log_entries

def write_json_test_case(output_dir, test_case_number, log_entries):
    """Writes a single test case to a .json file."""
    file_name = f"test_case_{test_case_number}.json"
    output_path = os.path.join(output_dir, file_name)
    
    with open(output_path, 'w') as json_file:
        json.dump(log_entries, json_file, indent=4)

def generate_test_cases(normal_logs, error_logs, num_test_cases, lower_bound, upper_bound, output_dir):
    """Generates multiple test case files with random log entries."""
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)
    
    for i in range(num_test_cases):
        # Randomize the total number of log lines for this test case
        total_logs = random.randint(lower_bound, upper_bound)
        
        # Randomly determine how many of the logs should be normal vs error
        normal_count = random.randint(0, total_logs)
        error_count = total_logs - normal_count
        
        # Create the test case log output
        log_entries = create_log_output(normal_logs, error_logs, normal_count, error_count)
        
        # Write the test case to a .json file
        write_json_test_case(output_dir, i + 1, log_entries)

def main():
    # Setup argument parser
    # EG) python3 multiple-testcase.py --num_test_cases 5 --lower_bound 5 --upper_bound 20
    # this will create 5 test cases (json format) with minimum of 5 loglines to max of 20 log lines
    # also could change the files for reading by argument
    parser = argparse.ArgumentParser(description='Generate test cases with log entries.')
    parser.add_argument('--num_test_cases', type=int, required=True, help='Number of test cases to generate')
    parser.add_argument('--lower_bound', type=int, required=True, help='Minimum number of logs per test case')
    parser.add_argument('--upper_bound', type=int, required=True, help='Maximum number of logs per test case')
    parser.add_argument('--output_dir', type=str, default='test_cases_json', help='Directory to save the test case files')
    parser.add_argument('--normal_file', type=str, default='normal_logs.txt', help='File with normal log lines')
    parser.add_argument('--error_file', type=str, default='error_logs.txt', help='File with error log lines')

    # Parse arguments
    args = parser.parse_args()

    # Read log files
    normal_logs = read_logs(args.normal_file)
    error_logs = read_logs(args.error_file)

    # Generate test cases
    generate_test_cases(normal_logs, error_logs, args.num_test_cases, args.lower_bound, args.upper_bound, args.output_dir)

    print(f"{args.num_test_cases} JSON test case files written to {args.output_dir}")

if __name__ == "__main__":
    main()
