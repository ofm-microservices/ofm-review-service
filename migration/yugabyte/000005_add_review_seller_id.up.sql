ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS seller_id UUID;

UPDATE reviews
SET seller_id = seller_user_id
WHERE seller_id IS NULL
  AND seller_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS reviews_seller_id_created_at_idx
	ON reviews (seller_id, created_at DESC, review_id DESC);

DROP INDEX IF EXISTS reviews_seller_user_created_at_idx;

ALTER TABLE reviews
DROP COLUMN IF EXISTS seller_user_id;
