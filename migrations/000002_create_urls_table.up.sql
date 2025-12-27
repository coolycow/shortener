CREATE TABLE urls (
                      id SERIAL PRIMARY KEY,
                      user_id INTEGER NOT NULL,
                      url TEXT NOT NULL,
                      key VARCHAR(255) NOT NULL
);

CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE UNIQUE INDEX idx_urls_user_id_url_unique ON urls(user_id, url);
CREATE UNIQUE INDEX idx_urls_key_unique ON urls(key);