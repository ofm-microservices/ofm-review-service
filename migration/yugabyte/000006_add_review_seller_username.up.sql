ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS seller_username TEXT;

UPDATE reviews
SET seller_username = COALESCE(seller_username, '')
WHERE seller_username IS NULL;

CREATE INDEX IF NOT EXISTS reviews_seller_username_created_at_idx
	ON reviews (seller_username, created_at DESC, review_id DESC);
