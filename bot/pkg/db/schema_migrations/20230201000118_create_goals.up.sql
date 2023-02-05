CREATE TABLE IF NOT EXISTS goals (
    id uuid DEFAULT uuid_generate_v4(),
    user_id uuid,
    goal int NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
)