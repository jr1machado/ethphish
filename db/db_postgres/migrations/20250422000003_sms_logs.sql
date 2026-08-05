-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS sms_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INTEGER,
    campaign_id INTEGER,
    r_id TEXT,
    send_date TIMESTAMP,
    send_attempt INTEGER,
    processing BOOLEAN,
    target TEXT
);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TABLE sms_logs;
