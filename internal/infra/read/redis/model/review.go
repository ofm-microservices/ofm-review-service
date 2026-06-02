package model

// ReviewCache is the Redis projection model for the review read model.
type ReviewCache struct {
	ID        string  `json:"review_id"`
	GigID     string  `json:"gig_id"`
	Content   string  `json:"content"`
	BuyerID   string  `json:"buyer_id"`
	Rating    int32   `json:"rating"`
	Author    *Author `json:"author,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// Author is the cached review author preview.
type Author struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	AvatarID    string `json:"avatar_id,omitempty"`
}

// RatingAggregate stores the review summary projection for a gig or seller.
type RatingAggregate struct {
	RatingAvg    float64 `json:"rating_avg"`
	TotalReviews int64   `json:"total_reviews"`
	Stars5       int64   `json:"stars_5"`
	Stars4       int64   `json:"stars_4"`
	Stars3       int64   `json:"stars_3"`
	Stars2       int64   `json:"stars_2"`
	Stars1       int64   `json:"stars_1"`
}
