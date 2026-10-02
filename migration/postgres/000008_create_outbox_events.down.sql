DROP TRIGGER IF EXISTS reviews_outbox ON reviews;
DROP FUNCTION IF EXISTS capture_review_outbox_event();
DROP TABLE IF EXISTS outbox_events;
