-- +goose Up
ALTER TABLE targets ADD COLUMN department VARCHAR(255);
ALTER TABLE targets ADD COLUMN company VARCHAR(255);
ALTER TABLE targets ADD COLUMN city VARCHAR(255);
ALTER TABLE targets ADD COLUMN state VARCHAR(255);
ALTER TABLE targets ADD COLUMN country VARCHAR(255);
ALTER TABLE targets ADD COLUMN unit VARCHAR(255);
ALTER TABLE targets ADD COLUMN tags TEXT;

ALTER TABLE results ADD COLUMN department VARCHAR(255);
ALTER TABLE results ADD COLUMN company VARCHAR(255);
ALTER TABLE results ADD COLUMN city VARCHAR(255);
ALTER TABLE results ADD COLUMN state VARCHAR(255);
ALTER TABLE results ADD COLUMN country VARCHAR(255);
ALTER TABLE results ADD COLUMN unit VARCHAR(255);
ALTER TABLE results ADD COLUMN tags TEXT;

ALTER TABLE email_requests ADD COLUMN department VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN company VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN city VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN state VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN country VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN unit VARCHAR(255);
ALTER TABLE email_requests ADD COLUMN tags TEXT;

ALTER TABLE sms_requests ADD COLUMN department VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN company VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN city VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN state VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN country VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN unit VARCHAR(255);
ALTER TABLE sms_requests ADD COLUMN tags TEXT;

-- +goose Down
ALTER TABLE targets DROP COLUMN department;
ALTER TABLE targets DROP COLUMN company;
ALTER TABLE targets DROP COLUMN city;
ALTER TABLE targets DROP COLUMN state;
ALTER TABLE targets DROP COLUMN country;
ALTER TABLE targets DROP COLUMN unit;
ALTER TABLE targets DROP COLUMN tags;

ALTER TABLE results DROP COLUMN department;
ALTER TABLE results DROP COLUMN company;
ALTER TABLE results DROP COLUMN city;
ALTER TABLE results DROP COLUMN state;
ALTER TABLE results DROP COLUMN country;
ALTER TABLE results DROP COLUMN unit;
ALTER TABLE results DROP COLUMN tags;

ALTER TABLE email_requests DROP COLUMN department;
ALTER TABLE email_requests DROP COLUMN company;
ALTER TABLE email_requests DROP COLUMN city;
ALTER TABLE email_requests DROP COLUMN state;
ALTER TABLE email_requests DROP COLUMN country;
ALTER TABLE email_requests DROP COLUMN unit;
ALTER TABLE email_requests DROP COLUMN tags;

ALTER TABLE sms_requests DROP COLUMN department;
ALTER TABLE sms_requests DROP COLUMN company;
ALTER TABLE sms_requests DROP COLUMN city;
ALTER TABLE sms_requests DROP COLUMN state;
ALTER TABLE sms_requests DROP COLUMN country;
ALTER TABLE sms_requests DROP COLUMN unit;
ALTER TABLE sms_requests DROP COLUMN tags;
