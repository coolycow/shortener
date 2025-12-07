CREATE TABLE urls (
                      id SERIAL PRIMARY KEY,
                      url TEXT NOT NULL,
                      key VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX idx_urls_url_unique ON urls(url);

CREATE UNIQUE INDEX idx_urls_key_unique ON urls(key);