import logging
import os
import sys
from pathlib import Path
from typing import Generator, Optional
import argparse

class LogReaderConfig:
    """Configuration class for log reader settings."""
    def __init__(self):
        self.log_file_path = os.getenv('LOG_FILE_PATH')
        self.batch_size = int(os.getenv('LOG_BATCH_SIZE', '1000'))

class LogReader:
    """Class to handle log file reading operations."""
    
    def __init__(self, config: LogReaderConfig):
        self.config = config
        self._setup_logging()
        self.logger = logging.getLogger(__name__)
        
    def _setup_logging(self) -> None:
        """Configure logging"""
        logging.basicConfig(
            format="%(asctime)s %(levelname)s [%(name)s] : %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
            handlers=[
                logging.StreamHandler(sys.stdout)
            ]
        )

    def validate_file_path(self) -> bool:
        """Validate the log file path exists and is readable."""
        if not self.config.log_file_path:
            self.logger.error("LOG_FILE_PATH environment variable is not set")
            return False
            
        file_path = Path(self.config.log_file_path)
        if not file_path.exists():
            self.logger.error(f"File does not exist: {file_path}")
            return False
        if not file_path.is_file():
            self.logger.error(f"Path is not a file: {file_path}")
            return False
        if not os.access(file_path, os.R_OK):
            self.logger.error(f"File is not readable: {file_path}")
            return False
            
        return True

    def read_log_file(self) -> Generator[str, None, None]:
        """
        Read log files and yield each line.
        
        Yields:
            str: Each line from the log file
        """
        if not self.validate_file_path():
            return

        self.logger.info(f"Starting to read log file: {self.config.log_file_path}")
        
        try:
            with open(self.config.log_file_path, 'r', encoding='utf-8') as log_file:
                for line_number, line in enumerate(log_file, 1):
                    try:
                        cleaned_line = line.strip()
                        if cleaned_line:  # Skip empty lines
                            yield cleaned_line
                        
                        # Log progress for large files
                        if line_number % self.config.batch_size == 0:
                            self.logger.info(f"Processed {line_number} lines")
                            
                    except UnicodeDecodeError as e:
                        self.logger.warning(f"Failed to decode line {line_number}: {e}")
                        continue
                        
        except Exception as e:
            self.logger.error(f"Error reading log file: {e}")
            raise

    def process_logs(self) -> Optional[int]:
        """
        Process the log file and return the number of lines processed.
        
        Returns:
            int: Number of lines processed, or None if processing failed
        """
        try:
            lines_processed = 0
            for line in self.read_log_file():
                # Add your log processing logic here
                self.logger.debug(f"Processing line: {line}")
                lines_processed += 1
                
            self.logger.info(f"Successfully processed {lines_processed} lines")
            return lines_processed
            
        except Exception as e:
            self.logger.error(f"Failed to process logs: {e}")
            return None

def main():
    parser = argparse.ArgumentParser(description='Process log files with configurable settings.')
    parser.add_argument('--log-file', help='Override LOG_FILE_PATH environment variable')
    args = parser.parse_args()

    config = LogReaderConfig()
    
    # Override config with command line arguments if provided
    if args.log_file:
        config.log_file_path = args.log_file

    # Initialize and run log reader
    log_reader = LogReader(config)
    result = log_reader.process_logs()
    
    sys.exit(0 if result is not None else 1)

if __name__ == "__main__":
    main()
