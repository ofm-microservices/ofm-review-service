DROP INDEX IF EXISTS reviews_seller_username_created_at_idx;

ALTER TABLE reviews
DROP COLUMN IF EXISTS seller_username;
