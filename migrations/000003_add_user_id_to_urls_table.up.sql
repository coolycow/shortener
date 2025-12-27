ALTER TABLE urls ADD COLUMN user_id INTEGER NOT NULL;

ALTER TABLE urls ADD CONSTRAINT fk_urls_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

DROP INDEX idx_urls_url_unique;

CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE UNIQUE INDEX idx_urls_user_id_url_unique ON urls(user_id, url);