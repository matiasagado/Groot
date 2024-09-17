So I want to be able to let this run for a long time. Not easy with SO API these days 10,000 a day with api key, 300 without

Maybe a docker stack, 

input query file -> python process getting question ids that match marks queries as done when in progess -> redis db 

(need to figure out how to deal with file locking. Maybe use a concept of a query file and directory mounted in volume)

another python process reading the redis ids to get and requesting the question bodies -> setting the data into another database
`
another python process querying all the answers and placing them in the db as well 

final python process reading from the database, querying openllama api to parse out the logs or errors  that relate to the input query  it would then write to a CSV on a per query basis


docker volume to hold the database. 



Should I gather all the question ids first, or 





id-fetcher    | Violation of backoff parameter (status code 400).
id-fetcher    | Status Code: 400
id-fetcher    | HTTP Error 400 received, but unable to extract wait time

..NEW..

user inputs query -> 

program listens to the queue, accepts them one at a time, adds the query to database and mark it as in progress. it will then get all the question ids and add them to a database related to the query. ->

program reads from this database and queries for all the question and answer data. Saves data to the database. Backs off api requests if getting an error. 

program that reads the database and makes queries to the LLM and refines the data, adds it to db and exports to a csv



Use redis queue to submit queries

use redis set to store ids that match the queries






