# Implementing Feature: Adding Classified Logs to ClickhouseDB

## Description
This features aims to add classified logs to clickhouseDB

## Table of Contents
- [Tasks](#Tasks)
- [ClickhouseDB](#ClickhouseDB)
- [Schema](#Schema)
- [Costs](#Costs)
- [Strategy](#Strategy)
- [Problems](#Problems)

## Tasks
- [x] Define Schema variables 
- [ ] Implement basic add
- [ ] Implement batched add


## ClickhouseDB
ClickHouse is a columnar database optimized for high-speed data ingestion, storage, and complex analytical queries. It’s particularly suitable for handling structured data and can manage high-throughput insertions efficiently, which aligns well with storing AI classification results

Key Benefits of ClickHouse for Classification Data:
- High Ingestion Speed: ClickHouse is optimized for rapid data insertion, which is ideal for streaming classified results.
- Efficient Storage and Compression: Its columnar storage and compression techniques help reduce storage costs.
- Analytical Querying: ClickHouse excels in analytical queries, enabling quick analysis of classifications for reporting and insights.

## Schema
```
CREATE TABLE IF NOT EXISTS vector_logs_experiment_2 (
			dt DateTime,
			file String,
			host String,
			level Nullable(String),
			user_defined String,
			original_message String,
			platform String,
			uuid UUID
		)
		ENGINE = MergeTree()
		PARTITION BY toYYYYMM(dt)
		ORDER BY (dt, uuid)
```
- dt: Timestamp of the log entry
- file: Log file source
- host: Host or server from which the log originated
- level: Log Level (INFO, ERROR, WARN)
- user_defined: ?
- original_message: Original log message
- platform: application source (nginx)
- uuid: Unique identifier for the log entry

## Costs

CPU Resources for Data Ingestion: While ClickHouse is efficient, the compute cost scales with the ingestion rate. Batch inserts or a stream ingestion strategy can optimize CPU usage.

Also, when adding the classified logs to the clickhouseDB
There is two methods:

1. Editing the existing ClickhouseDB

2. Creating a new clickhouse table for the classified results

## Strategy

>For high-frequency data insertion (e.g., real-time or near real-time), using batch inserts improves performance, as inserting one record at a time may reduce ClickHouse’s efficiency.

## Implement basic add

1. Add ai_classification to the schema

I want to keep it simple and have a boolean value on which it should be flagged or not

Updated schema:
```
CREATE TABLE IF NOT EXISTS vector_logs_experiment_2 (
			dt DateTime,
			file String,
			host String,
			level Nullable(String),
			user_defined String,
			original_message String,
            ai_classification UInt8,
			platform String,
			uuid UUID
		)
		ENGINE = MergeTree()
		PARTITION BY toYYYYMM(dt)
		ORDER BY (dt, uuid)
```

>Also, could add different fields like confidence. But, for now keeping it basic and focusing on the classification

To get the classification from the AI and send it to the clickhouseDB

There is two ways

1. ~~Send API ends from the ai-core.py to the pre-processor main.go~~ 
	- Erik said that it won't make sense to do this. As it would be better off rewriting ai-core.py into the main.go of the preprocessor if we need to.
2. Send the classifcation directly to ClickhouseDB

I will be implementing the 2nd option. However, in the future if needed ai-core.py could be rewritten into main.go in preprocessor. 

The reason why rewriting ai-core.py into main.go is centralizing interaction with clickhouseDB into one file. (might be more reasons that I do not know of)

## 1st Draft Pull Request

While, I was sending the classified data to clickhouseDb. I realized I would need to compare UUID so that it will find the right entry to edit the processed logs.

Currently, ran into a problem of passing UUID from and to GO and python. UUID is 16byte integer which isn't supported by JSON format. 

I am sending JSON format between ai-core.py and pre-processor main.go. However, when I send UUID between these two in JSON format UUID outputs are all 0s.
```
Received log UUID: [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
poc-log-preprocessor    | Processing AI-processed log with UUID: 00000000-0000-0000-0000-000000000000
poc-log-preprocessor    | Full log received: {Dt:2024-10-31 08:27:52 +0000 UTC File:/var/log/nginx/access.log Host:9a6018fd0fb0 Level:<nil> UserDefined:map[nginx:map[agent:Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko client:46.174.191.32 method:GET path:/ protocol:HTTP/1.0 request:GET / HTTP/1.0 size:35 status:444 user:<nil>]] OriginalMessage:46.174.191.32 - - [31/Oct/2024:01:27:52 -0700] "GET / HTTP/1.0" 444 35 "-" "Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko" Platform:Nginx UUID:[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]}
poc-log-preprocessor    | User-defined data for UUID 00000000-0000-0000-0000-000000000000: {"nginx":{"agent":"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko","client":"46.174.191.32","method":"GET","path":"/","protocol":"HTTP/1.0","request":"GET / HTTP/1.0","size":35,"status":444,"user":null}}
poc-log-preprocessor    | AI classification for UUID 00000000-0000-0000-0000-000000000000: 0
poc-log-preprocessor    | {"time":"2024-11-08T06:23:53.052714652Z","level":"ERROR","prefix":"echo","file":"main.go","line":"355","message":"Failed to process AI log with UUID \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00: failed to update ClickHouse record: DB::Exception: Unknown function JSON_MERGE_PATCH. Maybe you meant: ['JSONMergePatch','jsonMergePatch']: While processing _CAST(if(uuid = '00000000-0000-0000-0000-000000000000', _CAST(JSON_MERGE_PATCH(user_defined, '{\"nginx\":{\"agent\":\"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko\",\"client\":\"46.174.191.32\",\"method\":\"GET\",\"path\":\"/\",\"protocol\":\"HTTP/1.0\",\"request\":\"GET / HTTP/1.0\",\"size\":35,\"status\":444,\"user\":null}}'), 'String'), user_defined), 'String')"}
```
As seen above, I am getting error because the UUID is all 0s and the it can't find the right entry to edit in clickhouse. 

Am I going down the right path currently? and also would it be the right idea to encode the UUID so that it will be able to be sent in JSON format. I believe the reason why UUID is all 0s is because JSON format doesn't support 16byte integers.

Realized that I was going through log preprocessor to update clickhouseDB. Rolled back the changes done so its easier to document.

Will do more commits with concise messages to show what work I've done so it's easier to track workflow

## Nov/11/2024

Commit: ```85ca833ff17def53cb39c61b376877895b3f2241```

Rolled back pro-processor main.go to the original master.

I added requirements inside ai-core
clickhouse-driver
for interacting with clickhouse databse

## Nov/18/2024

Updated pre-processor so that UUID is type string

Included migrations on goose for schema management
Currently, only has ai_classificationlevel

First, docker-compose build the project
than apply the schema through goose
Rebuilld the docker-compose again with the applied schema changes

How to apply schema changes on goose:
```
goose -dir migrations clickhouse "clickhouse://default:password@localhost:9000/default" up
```
Checking status of the migrations
```
goose -dir migrations clickhouse "clickhouse://default:password@localhost:9000/default" status
```
Reverting applied changes
```
goose -dir migrations clickhouse "clickhouse://default:password@localhost:9000/default" down
```

## Nov/19/2024

Why UUID string?

Instead of using the UUID type in clickhouseDB. I decided to use the string type. The main reason is ai-core is written in python. Which python use UUID as string. Which means if UUID is stored as a 16byte int array we would need to process a conversion within ai-core.py to change the int array into the string representation to process. Which is needless complexity at the current state of the project. Hence, UUID will be stored as string. However, in the future this might change depending on how much storage efficiency matter. On the perspective of performance, the lookups are probably mostly going to be done by time stamps instead of UUID. Showing more proof of reasons to use UUID as type string.

Set up Goose migrations through docker-compose.yaml

## Nov/20/2024

Updated docker-compose.yaml so that it will take migrations through goose automatically by using goose

## Nov/21/2024

Fixed the bug where clickhouse table wasn't correctly setup or ai classification not being added properly.
```
volumes:
	- ./container-data/redis/data:/data
```
Had to do with Redis keeping states between runs. Which was causing weird behavior.

Whether UUID is good with ClickhouseDB?

UUID in ClickhouseDB are sorted by the second half of the UUID. When data is sorted by the second half of the UUID, it could introduce fragmentation in the data storage layout. While, also leading to poor insert performance. Overall, the only reason UUID would be useful for ClickhouseDB is the uniqueness. In the perspective of large incoming log data, UUID might be useful. However, for high-performance query operations it might be better idea to use simple integer index or other variables.

## Nov/23/2024

Fixed the path file in docker-compose for migrations

Checking ClickhouseDB

Terminal:
```
docker exec -it poc-clickhouse clickhouse-client --user=default --password=password --database=default
```
Clickhouse Server:
```
SELECT uuid, ai_classified_level, level
FROM user_log_data
LIMIT 10;
```
