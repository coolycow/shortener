CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE urls (
                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                      user_id UUID NOT NULL,
                      url TEXT NOT NULL,
                      key VARCHAR(255) NOT NULL,
                      created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                      updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                      deleted_at TIMESTAMP WITH TIME ZONE NULL DEFAULT NULL
);

CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE UNIQUE INDEX idx_urls_user_id_url_unique ON urls(user_id, url);
CREATE UNIQUE INDEX idx_urls_key_unique ON urls(key);