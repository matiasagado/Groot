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

def calculate_metrics(true_labels, predicted_labels):
    """Calculate classification performance metrics."""
    accuracy = accuracy_score(true_labels, predicted_labels)
    precision = precision_score(true_labels, predicted_labels, pos_label='ERROR')
    recall = recall_score(true_labels, predicted_labels, pos_label='ERROR')
    f1 = f1_score(true_labels, predicted_labels, pos_label='ERROR')
    conf_matrix = confusion_matrix(true_labels, predicted_labels)
    return accuracy, precision, recall, f1, conf_matrix

def display_metrics(accuracy, precision, recall, f1, conf_matrix):
    """Display the performance metrics."""
    print(f"Accuracy: {accuracy:.2f}")
    print(f"Precision: {precision:.2f}")
    print(f"Recall: {recall:.2f}")
    print(f"F1 Score: {f1:.2f}")
    print("Confusion Matrix:")
    print(conf_matrix)

def main():

    #Step 1: Load the classification data
    data = load_classification_data(json_file_path)

    if data:

        # Step 2: Extract labels
        true_labels, predicted_lables = extract_labels(data)

        # Step 3: Calculate metrics
        accuracy, precision, recall, f1, conf_matrix = calculate_metrics(true_labels, predicted_labels)

        # Step 4: Display the metrics
        display_metrics(accuracy, precision, recall, f1, conf_matrix)

if __name__ == '__main__':
    main()