import json
from sklearn.metrics import accoruacy_score, precision_score, recall_score, f1_score, confusion_matrix
import os

# define the paths
json_file_path = os.path.join(os.getcwd(), '../../backend/src/classification_files.json')

if __name__ == '__main__':
    main()