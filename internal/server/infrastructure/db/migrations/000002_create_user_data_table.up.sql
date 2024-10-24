BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS user_data(
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    info_type VARCHAR(50) NOT NULL CHECK (info_type IN ('login_password', 'text', 'binary', 'bank_card')),
    info TEXT,
    meta TEXT,
    created TIMESTAMP
);

COMMIT;
