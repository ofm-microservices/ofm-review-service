CREATE TABLE reviews (
    review_id UUID PRIMARY KEY,
    order_id UUID NOT NULL,
    gig_id UUID NOT NULL,
    content TEXT NOT NULL,
    buyer_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX reviews_order_id_idx ON reviews (order_id);
CREATE UNIQUE INDEX reviews_buyer_gig_idx ON reviews (buyer_id, gig_id);
