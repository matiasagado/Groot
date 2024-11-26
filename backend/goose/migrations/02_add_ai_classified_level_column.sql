-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_log_data
ADD COLUMN ai_classified_level String;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_log_data
DROP COLUMN ai_classified_level;
-- +goose StatementEnd
