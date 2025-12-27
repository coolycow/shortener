DROP INDEX IF EXISTS idx_urls_user_id;
DROP INDEX IF EXISTS idx_urls_user_id_url_unique;
ALTER TABLE urls DROP CONSTRAINT IF EXISTS fk_urls_user_id;
ALTER TABLE urls DROP COLUMN user_id;
CREATE UNIQUE INDEX idx_urls_url_unique ON urls(url);