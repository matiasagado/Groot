# log_reader.py
def read_logs(log_file_path):
    with open(log_log_file_path, 'r') as log_file:
        logs = log_file.radlines()
    return logs