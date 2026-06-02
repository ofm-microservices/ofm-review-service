package repository

const (
	insertReviewQuery = `
		INSERT INTO reviews (review_id, order_id, gig_id, content, buyer_id, seller_id, rating)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
	`

	selectReviewByIDQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE review_id = $1
	`

	selectReviewByOrderIDQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE order_id = $1
	`

	selectReviewsByGigIDQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE gig_id = $1
		ORDER BY created_at DESC, review_id DESC
		LIMIT $2
	`

	selectReviewsByGigIDCursorQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE gig_id = $1
		  AND (created_at, review_id) < ($2, $3)
		ORDER BY created_at DESC, review_id DESC
		LIMIT $4
	`

	selectReviewsBySellerIDQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE seller_id = $1
		ORDER BY created_at DESC, review_id DESC
		LIMIT $2
	`

	selectReviewsBySellerIDCursorQuery = `
		SELECT review_id, order_id, gig_id, content, buyer_id, seller_id, rating, created_at, updated_at
		FROM reviews
		WHERE seller_id = $1
		  AND (created_at, review_id) < ($2, $3)
		ORDER BY created_at DESC, review_id DESC
		LIMIT $4
	`

	selectGigRatingSummaryQuery = `
		SELECT
			COALESCE(AVG(rating), 0) AS rating_avg,
			COUNT(*) AS total_reviews,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) AS stars_5,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) AS stars_4,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) AS stars_3,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) AS stars_2,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) AS stars_1
		FROM reviews
		WHERE gig_id = $1
	`

	selectSellerRatingSummaryQuery = `
		SELECT
			COALESCE(AVG(rating), 0) AS rating_avg,
			COUNT(*) AS total_reviews,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) AS stars_5,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) AS stars_4,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) AS stars_3,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) AS stars_2,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) AS stars_1
		FROM reviews
		WHERE seller_id = $1
	`

	selectGigRatingSummariesQuery = `
		SELECT
			gig_id AS owner_id,
			COALESCE(AVG(rating), 0) AS rating_avg,
			COUNT(*) AS total_reviews,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) AS stars_5,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) AS stars_4,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) AS stars_3,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) AS stars_2,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) AS stars_1
		FROM reviews
		GROUP BY gig_id
		ORDER BY gig_id
	`

	selectSellerRatingSummariesQuery = `
		SELECT
			seller_id AS owner_id,
			COALESCE(AVG(rating), 0) AS rating_avg,
			COUNT(*) AS total_reviews,
			COALESCE(SUM(CASE WHEN rating = 5 THEN 1 ELSE 0 END), 0) AS stars_5,
			COALESCE(SUM(CASE WHEN rating = 4 THEN 1 ELSE 0 END), 0) AS stars_4,
			COALESCE(SUM(CASE WHEN rating = 3 THEN 1 ELSE 0 END), 0) AS stars_3,
			COALESCE(SUM(CASE WHEN rating = 2 THEN 1 ELSE 0 END), 0) AS stars_2,
			COALESCE(SUM(CASE WHEN rating = 1 THEN 1 ELSE 0 END), 0) AS stars_1
		FROM reviews
		GROUP BY seller_id
		ORDER BY seller_id
	`

	deleteReviewByIDQuery = `
		DELETE FROM reviews
		WHERE review_id = $1
	`
)
