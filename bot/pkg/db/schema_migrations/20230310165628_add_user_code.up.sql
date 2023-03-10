CREATE TABLE IF NOT EXISTS user_code (
     id uuid DEFAULT uuid_generate_v4(),
     user_id uuid NOT NULL,
     code text NOT NULL UNIQUE,
     valid boolean NOT NULL DEFAULT true,
     expired_at timestamp with time zone NOT NULL DEFAULT now() + interval '1 hour',
     created_at timestamp with time zone DEFAULT now(),
     updated_at timestamp with time zone DEFAULT now(),
     PRIMARY KEY (id),
     CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)
);