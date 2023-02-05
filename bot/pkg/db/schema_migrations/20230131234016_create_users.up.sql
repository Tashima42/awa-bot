CREATE TABLE IF NOT EXISTS users(
    id uuid DEFAULT uuid_generate_v4(),
    telegram_id int UNIQUE NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id)
);
