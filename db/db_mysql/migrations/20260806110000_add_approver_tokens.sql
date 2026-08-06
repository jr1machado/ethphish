-- +goose Up
ALTER TABLE contract_approvers ADD COLUMN token_hash varchar(255);
ALTER TABLE contract_approvers ADD COLUMN token_expires_at datetime;
ALTER TABLE contract_approvers ADD COLUMN active_approval_request_id integer;

CREATE INDEX idx_contract_approvers_token_hash ON contract_approvers(token_hash);

-- +goose Down
DROP INDEX idx_contract_approvers_token_hash ON contract_approvers;
ALTER TABLE contract_approvers DROP COLUMN token_hash;
ALTER TABLE contract_approvers DROP COLUMN token_expires_at;
ALTER TABLE contract_approvers DROP COLUMN active_approval_request_id;
