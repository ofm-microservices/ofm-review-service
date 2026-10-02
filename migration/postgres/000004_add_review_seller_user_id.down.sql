DROP INDEX IF EXISTS reviews_seller_user_created_at_idx;

ALTER TABLE reviews
DROP COLUMN IF EXISTS seller_user_id;
