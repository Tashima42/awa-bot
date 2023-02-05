CREATE TABLE IF NOT EXISTS competition_users(
    id uuid DEFAULT uuid_generate_v4(),
    user_id uuid,
    competition_id uuid,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_competition FOREIGN KEY (competition_id) REFERENCES competition(id)
);
