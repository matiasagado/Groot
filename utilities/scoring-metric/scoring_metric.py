import json
from sklearn.metrics import accoruacy_score, precision_score, recall_score, f1_score, confusion_matrix
import os

# define the paths
json_file_path = os.path.join(os.getcwd(), '../../backend/src/classification_files.json')

def load_classification_data(file_path):
    """Load classification data from JSON file"""
    try:
        with open(json_path, 'r') as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        print(f"File not found: {json_path}")
        return []
    except json.JSONDecodeError:
        print(f"JSON decoding error: {json_path}")
        return []
    except Exception:
        print(f"Exception")
        return []

def extract_labels(data):
    """Extract labels from the JSON data"""
    true_labels = [entry['classification'] for entry in data]
    # TODO: Replace this with actual LLM predicted labels
    predicted_labels = ['ERROR' if 'error' in entry['log_line'].lower() else 'INFO' for entry in data]
    return true_labels, predicted_labels

def main():

    #Step 1: Load the classification data
    data = load_classification_data(json_file_path)

if __name__ == '__main__':
    main()