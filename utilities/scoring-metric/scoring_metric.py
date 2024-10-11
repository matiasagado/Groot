import json
from sklearn.metrics import accuracy_score, precision_score, recall_score, f1_score, confusion_matrix
import os
import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

# Define the paths (adjust the path according to your file structure)
json_file_path = os.path.join(os.getcwd(), '../../backend/src/classification_files.json')

def load_classification_data(file_path):
    """
    Load classification data from the provided JSON file.
    Returns:
        - data: A list of dictionaries, each containing 'classification' and 'log_line'.
    """
    try:
        # Open the file and load the JSON content
        with open(file_path, 'r') as file:
            data = json.load(file)
            return data
    except FileNotFoundError:
        # Handle case when the file is not found
        print(f"File not found: {file_path}")
        return []
    except json.JSONDecodeError:
        # Handle case when there is an issue parsing the JSON
        print(f"JSON decoding error: {file_path}")
        return []
    except Exception as e:
        # Catch any other exceptions
        print(f"An unexpected error occurred: {str(e)}")
        return []

def extract_labels(data):
    """
    Extract true labels and predicted labels from the JSON data.
    True labels are from the 'classification' key.
    Predicted labels are inferred based on the content of 'log_line'.
    Returns:
        - true_labels: List of true classifications from the data.
        - predicted_labels: List of predicted classifications based on log content.
    """
    # Extract true labels from the JSON
    true_labels = [entry['classification'] for entry in data]
    
    # Use basic rules to predict labels based on log content ('ERROR' or 'INFO')
    predicted_labels = ['ERROR' if 'error' in entry['log_line'].lower() else 'INFO' for entry in data]
    
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
    # Calculate accuracy (correct predictions / total predictions)
    accuracy = accuracy_score(true_labels, predicted_labels)
    
    # Calculate precision for 'ERROR' label
    precision = precision_score(true_labels, predicted_labels, pos_label='ERROR')
    
    # Calculate recall for 'ERROR' label
    recall = recall_score(true_labels, predicted_labels, pos_label='ERROR')
    
    # Calculate F1 score for 'ERROR' label
    f1 = f1_score(true_labels, predicted_labels, pos_label='ERROR')
    
    # Generate confusion matrix for 'ERROR' and 'INFO'
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
    sns.heatmap(conf_matrix, annot=True, fmt='d', cmap='Blues')
    plt.title('Confusion Matrix')
    plt.ylabel('True Labels')
    plt.xlabel('Predicted Labels')
    plt.show()

def save_results(true_labels, predicted_labels, accuracy, precision, recall, f1, conf_matrix):
    """
    Save classification results and metrics to a CSV file.
    Logs true labels, predicted labels, and metrics in a pandas DataFrame.
    """
    # Create a pandas DataFrame for saving the results
    df = pd.DataFrame({
        'True Labels': true_labels,
        'Predicted Labels': predicted_labels
    })
    
    # Append metrics as a summary at the bottom of the CSV file
    summary_df = pd.DataFrame({
        'Metric': ['Accuracy', 'Precision', 'Recall', 'F1 Score'],
        'Score': [accuracy, precision, recall, f1]
    })
    
    # Save the results and summary to a CSV file
    df.to_csv('classification_results.csv', index=False)
    summary_df.to_csv('classification_results.csv', mode='a', header=True, index=False)

def main():
    """
    Main function that drives the log classification and performance evaluation.
    Steps:
        1. Load the classification data.
        2. Extract true and predicted labels.
        3. Calculate performance metrics.
        4. Display and plot the metrics.
        5. Save the results to a CSV file.
    """
    # Step 1: Load the classification data
    data = load_classification_data(json_file_path)

    if data:
        # Step 2: Extract true and predicted labels
        true_labels, predicted_labels = extract_labels(data)

        # Step 3: Calculate performance metrics
        accuracy, precision, recall, f1, conf_matrix = calculate_metrics(true_labels, predicted_labels)

        # Step 4: Display the metrics and plot the confusion matrix
        display_metrics(accuracy, precision, recall, f1, conf_matrix)
        plot_confusion_matrix(conf_matrix)

        # Step 5: Save the results to a CSV file
        save_results(true_labels, predicted_labels, accuracy, precision, recall, f1, conf_matrix)

if __name__ == '__main__':
    main()
