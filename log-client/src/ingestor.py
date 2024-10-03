import logging
import os

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s : %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S",
)

# Example usage
logger = logging.getLogger(__name__)

# Method to read raw log files from servers
# For now takes in a txt file and 
def read_log_file(file_path):
    """Reads log files and yields each line."""
    logger.info("Reading log files")
    logger.info("test")
    if not os.path.exists(file_path):
        logger.error(f"File {file_path} does not exist.")
        return

    with open(file_path, 'r') as log_file:
        logger.info("Opened log files")
        for line in log_file:
            print(line)
            #yield line.strip()  # Yield each log line without leading/trailing spaces

if __name__ == "__main__":
    log_file_path = '/path/to/your/log/file.log'  # Change this to your log file path
    for log_line in read_log_file(log_file_path):
        logger.info(f"Log line: {log_line}")