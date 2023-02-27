CREATE TABLE IF NOT EXISTS auth(
    id uuid DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL UNIQUE,
    api_key text NOT NULL UNIQUE,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
)