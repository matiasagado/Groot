import random
import json

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

def write_json_output(output_path, log_entries):
    """Writes the log entries to a JSON file."""
    with open(output_path, 'w') as json_file:
        json.dump(log_entries, json_file, indent=4)

def main(normal_file, error_file, output_file, normal_count, error_count):
    # Read log files
    normal_logs = read_logs(normal_file)
    error_logs = read_logs(error_file)

    # Create the log output
    log_output = create_log_output(normal_logs, error_logs, normal_count, error_count)

    # Write to JSON
    write_json_output(output_file, log_output)
    print(f"Logs written to {output_file}")

if __name__ == "__main__":
    # Input and output file paths
    normal_file = "normal_logs.txt"
    error_file = "error_logs.txt"
    output_file = "logs_output.json"
    
    # Number of normal and error log lines to include
    normal_count = 5  # Modify this as needed
    error_count = 1    # Modify this as needed

    # Run the main function
    main(normal_file, error_file, output_file, normal_count, error_count)
