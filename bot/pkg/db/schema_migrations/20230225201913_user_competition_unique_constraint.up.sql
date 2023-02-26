ALTER TABLE IF EXISTS competition_users
    ADD CONSTRAINT unique_user_competition UNIQUE (user_id, competition_id);