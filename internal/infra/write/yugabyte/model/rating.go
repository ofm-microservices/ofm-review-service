package model

// RatingSummaryRow is the grouped Yugabyte row used to seed a gig or seller
// rating summary into Redis.
type RatingSummaryRow struct {
	OwnerID      string  `db:"owner_id"`
	RatingAvg    float64 `db:"rating_avg"`
	TotalReviews int64   `db:"total_reviews"`
	Stars5       int64   `db:"stars_5"`
	Stars4       int64   `db:"stars_4"`
	Stars3       int64   `db:"stars_3"`
	Stars2       int64   `db:"stars_2"`
	Stars1       int64   `db:"stars_1"`
}
