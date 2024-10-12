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