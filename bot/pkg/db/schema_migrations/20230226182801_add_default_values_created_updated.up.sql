BEGIN;
    -- users
    ALTER TABLE IF EXISTS users
        ALTER COLUMN created_at SET DEFAULT now();
    ALTER TABLE IF EXISTS users
        ALTER COLUMN updated_at SET DEFAULT now();
    -- water
    ALTER TABLE IF EXISTS users
        ALTER COLUMN created_at SET DEFAULT now();
    ALTER TABLE IF EXISTS users
        ALTER COLUMN updated_at SET DEFAULT now();
    -- competition
    ALTER TABLE IF EXISTS competition
        ALTER COLUMN created_at SET DEFAULT now();
    ALTER TABLE IF EXISTS competition
        ALTER COLUMN updated_at SET DEFAULT now();
    -- goals
    ALTER TABLE IF EXISTS goals
        ALTER COLUMN created_at SET DEFAULT now();
    ALTER TABLE IF EXISTS goals
        ALTER COLUMN updated_at SET DEFAULT now();
    -- competition_users
    ALTER TABLE IF EXISTS competition_users
        ALTER COLUMN created_at SET DEFAULT now();
    ALTER TABLE IF EXISTS competition_users
        ALTER COLUMN updated_at SET DEFAULT now();
COMMIT;
