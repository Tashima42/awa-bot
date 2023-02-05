CREATE TABLE IF NOT EXISTS water(
    id uuid DEFAULT uuid_generate_v4(),
    user_id uuid,
    amount int NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);
