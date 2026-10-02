package model

import (
	"database/sql"
	"time"
)

// ReviewRow is the PostgreSQL persistence model for the review write model.
type ReviewRow struct {
	ID             string         `db:"review_id"`
	OrderID        string         `db:"order_id"`
	GigID          string         `db:"gig_id"`
	Content        string         `db:"content"`
	BuyerID        string         `db:"buyer_id"`
	BuyerUsername  sql.NullString `db:"buyer_username"`
	SellerID       sql.NullString `db:"seller_id"`
	SellerUsername sql.NullString `db:"seller_username"`
	Rating         int32          `db:"rating"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}
