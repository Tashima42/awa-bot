BEGIN;
    -- users
    ALTER TABLE IF EXISTS users
        ALTER COLUMN created_at DROP DEFAULT;
    ALTER TABLE IF EXISTS users
        ALTER COLUMN updated_at DROP DEFAULT;
    -- water
    ALTER TABLE IF EXISTS users
        ALTER COLUMN created_at DROP DEFAULT;
    ALTER TABLE IF EXISTS users
        ALTER COLUMN updated_at DROP DEFAULT;
    -- competition
    ALTER TABLE IF EXISTS competition
        ALTER COLUMN created_at DROP DEFAULT;
    ALTER TABLE IF EXISTS competition
        ALTER COLUMN updated_at DROP DEFAULT;
    -- goals
    ALTER TABLE IF EXISTS goals
        ALTER COLUMN created_at DROP DEFAULT;
    ALTER TABLE IF EXISTS goals
        ALTER COLUMN updated_at DROP DEFAULT;
    -- competition_users
    ALTER TABLE IF EXISTS competition_users
        ALTER COLUMN created_at DROP DEFAULT;
    ALTER TABLE IF EXISTS competition_users
        ALTER COLUMN updated_at DROP DEFAULT;
COMMIT;
