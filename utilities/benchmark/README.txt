Build docker image:
docker build -t benchmarker .

Run the image:
docker run -it --name benchmarker-container benchmarker

Copy the data from the container:
docker cp benchmarker-container:/app/test_cases_json [PATH TO COPY TO]
[PATH TO COPY TO] = Replace with whatever path to paste the testcases directory to

Updating Container:
docker stop benchmarker-container
docker rm benchmarker-container

Tried to incorporate dockerfile to run everything in a isolated environment, but ran into some problems
So, decided to run it through locally trying to build a workflow that will apply to huge dataset we will be using the benchmarking tool to do.

First, the multiple-testcase.py will generate test cases based on two test files. The two text files are from normal log lines given to us and a error log lines text file from stack over flow posts. Currently, the generator will grab random log lines from the two files. However, this will be changed later on to chunk the files into a manageable pieces for the AI to proccess.

The input: 
  EG) python3 multiple-test_case.py --num_test_cases 5 --lower_bound 5 --upper_bound 20
  This will create 5 test cases (json format) with minimum of 5 loglines to max of 20 log lines
  also could change the files for reading by argument

The output of the multiple-testcase.py would be the json files specified by the user inside the test_cases_json directory

After, running the test case generator the ai-core.py will run with input command:
python3 ai-core.py
The ai-core.py will read the
Directory containing the test case files
    test_case_directory = "test_cases_json"

The output: json files will be saved to the output_directory in classified_test_cases

Then the scoring.py will be execuated with command:
python3 scoring.py test_cases_json classified_test_cases

Which will read the true and predicted labels from the two json file and output a graph or data. If the argument is a directory
it will aggregarte all the results from the different test cases into one score.


