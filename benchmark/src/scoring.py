import json
import os
import argparse
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score, confusion_matrix

def load_classification_data(file_path):
    """
    Load true classification data from the provided JSON file.
    Returns:
        - data: A list of dictionaries, each containing 'log' and 'is_normal' keys.
    """
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return []

    try:
        with open(file_path, 'r') as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        print(f"File not found: {file_path}")
        return []
    except json.JSONDecodeError:
        print(f"JSON decoding error: {file_path}")
        return []
    except Exception as e:
        print(f"An unexpected error occurred: {str(e)}")
        return []

def load_llama_classifications(file_path):
    """
    Load LLama's classification data from the provided JSON file.
    Returns:
        - data: A list of dictionaries, each containing 'log_line' and 'classification'.
    """
    if not os.path.exists(file_path):
        print(f"File not found: {file_path}")
        return []

    try:
        with open(file_path, 'r') as file:
            llama_data = json.load(file)
            return llama_data
    except FileNotFoundError:
        print(f"File not found: {file_path}")
        return []
    except json.JSONDecodeError:
        print(f"JSON decoding error: {file_path}")
        return []
    except Exception as e:
        print(f"An unexpected error occurred: {str(e)}")
        return []

def extract_labels_and_predictions(true_data, llama_data):
    """
    Extract true labels and LLama's predicted labels from the JSON data.
    True labels are inferred from the boolean 'is_normal' key.
    LLama's predicted labels are extracted from its output.
    Returns:
        - true_labels: List of true classifications from the data ('INFO' or 'ERROR').
        - predicted_labels: List of classifications by LLama.
    """
    # Extract true labels from the boolean values (True -> INFO, False -> ERROR)
    true_labels = ['INFO' if entry['is_normal'] else 'ERROR' for entry in true_data]
    
    # Extract predicted labels from LLama's output
    predicted_labels = [entry['classification'] for entry in llama_data]
    
    return true_labels, predicted_labels

def calculate_metrics(true_labels, predicted_labels):
    """
    Calculate classification performance metrics.
    Metrics include accuracy, precision, recall, F1 score, and confusion matrix.
    Returns:
        - accuracy: The overall accuracy of the classification.
        - precision: The precision for classifying 'ERROR'.
        - recall: The recall for classifying 'ERROR'.
        - f1: The F1 score, a balance between precision and recall.
        - conf_matrix: A confusion matrix to summarize the classification performance.
    """
    accuracy = accuracy_score(true_labels, predicted_labels)
    precision = precision_score(true_labels, predicted_labels, pos_label='ERROR')
    recall = recall_score(true_labels, predicted_labels, pos_label='ERROR')
    f1 = f1_score(true_labels, predicted_labels, pos_label='ERROR')
    conf_matrix = confusion_matrix(true_labels, predicted_labels)
    
    return accuracy, precision, recall, f1, conf_matrix

def display_metrics(accuracy, precision, recall, f1, conf_matrix):
    """
    Display the calculated metrics in a readable format.
    Prints accuracy, precision, recall, F1 score, and confusion matrix.
    """
    print(f"Accuracy: {accuracy:.2f}")
    print(f"Precision: {precision:.2f}")
    print(f"Recall: {recall:.2f}")
    print(f"F1 Score: {f1:.2f}")
    print("Confusion Matrix:")
    print(conf_matrix)

def plot_confusion_matrix(conf_matrix):
    """
    Plot the confusion matrix using Seaborn for visualization.
    Displays the confusion matrix as a heatmap.
    """
    sns.heatmap(conf_matrix, annot=True, fmt='d', cmap='Blues', xticklabels=['INFO', 'ERROR'], yticklabels=['INFO', 'ERROR'])
    plt.title('Confusion Matrix')
    plt.ylabel('True Labels')
    plt.xlabel('Predicted Labels')
    plt.show()

def save_results(true_labels, predicted_labels, accuracy, precision, recall, f1, conf_matrix):
    """
    Save classification results and metrics to a JSON file.
    Logs true labels, predicted labels, and metrics in a dictionary format and writes to a JSON file.
    """
    results = {
        'results': [
            {'True Label': true, 'Predicted Label': pred} for true, pred in zip(true_labels, predicted_labels)
        ],
        'metrics': {
            'Accuracy': accuracy,
            'Precision': precision,
            'Recall': recall,
            'F1 Score': f1,
            'Confusion Matrix': conf_matrix.tolist()  # Convert to list for JSON serialization
        }
    }

    with open('metric_results.json', 'w') as json_file:
        json.dump(results, json_file, indent=4)

def main(true_file, llama_file):
    """
    Main function to drive the log classification and performance evaluation using LLama predictions.
    """
    # Load the true classification data (log with boolean values)
    true_data = load_classification_data(true_file)

    # Load LLama's predicted classification data
    llama_data = load_llama_classifications(llama_file)

    if true_data and llama_data:
        # Extract true and LLama's predicted labels
        true_labels, predicted_labels = extract_labels_and_predictions(true_data, llama_data)

        # Calculate performance metrics
        accuracy, precision, recall, f1, conf_matrix = calculate_metrics(true_labels, predicted_labels)

        # Display and plot the metrics
        display_metrics(accuracy, precision, recall, f1, conf_matrix)
        plot_confusion_matrix(conf_matrix)

        # Save the results to a JSON file
        save_results(true_labels, predicted_labels, accuracy, precision, recall, f1, conf_matrix)

if __name__ == '__main__':
    # Argument parser to take in the file paths
    parser = argparse.ArgumentParser(description="Evaluate LLama's log classification performance.")
    parser.add_argument('true_file', type=str, help="Path to the JSON file with true labels.")
    parser.add_argument('llama_file', type=str, help="Path to the JSON file with LLama's predicted classifications.")

    args = parser.parse_args()
    
    main(args.true_file, args.llama_file)
