DROP INDEX IF EXISTS reviews_buyer_order_idx;
CREATE UNIQUE INDEX reviews_buyer_gig_idx ON reviews (buyer_id, gig_id);
