CREATE TABLE IF NOT EXISTS competition(
    id uuid DEFAULT uuid_generate_v4(),
    chat_id int UNIQUE NOT NULL,
    start_date timestamp NOT NULL,
    end_date timestamp NOT NULL,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id)
);