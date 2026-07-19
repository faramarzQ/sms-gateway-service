CREATE TABLE users
(
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    traffic_class VARCHAR(20) NOT NULL DEFAULT 'standard',
    balance BIGINT NOT NULL,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);

CREATE INDEX idx_users_deleted_at
    ON users(deleted_at);