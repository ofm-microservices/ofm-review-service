ALTER TABLE reviews
ADD COLUMN IF NOT EXISTS buyer_username TEXT;

UPDATE reviews
SET buyer_username = COALESCE(buyer_username, '')
WHERE buyer_username IS NULL;
