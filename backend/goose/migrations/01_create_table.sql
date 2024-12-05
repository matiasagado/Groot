-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_log_data (
			dt DateTime,
			file String,
			host String,
			level Nullable(String),
			user_defined String,
			original_message String,
			platform String,
			uuid String
		)
		ENGINE = MergeTree()
		PARTITION BY toYYYYMM(dt)
		ORDER BY (dt, uuid)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_log_data;
-- +goose StatementEnd
