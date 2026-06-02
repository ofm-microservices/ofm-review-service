DROP INDEX IF EXISTS reviews_seller_id_created_at_idx;

ALTER TABLE reviews
DROP COLUMN IF EXISTS seller_id;

ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS seller_user_id UUID;

CREATE INDEX IF NOT EXISTS reviews_seller_user_created_at_idx
	ON reviews (seller_user_id, created_at DESC, review_id DESC);
