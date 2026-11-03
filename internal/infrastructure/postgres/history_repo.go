package postgres

import (
	"context"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HistoryRepo struct {
	pool *pgxpool.Pool
}

func NewHistoryRepo(pool *pgxpool.Pool) *HistoryRepo {
	return &HistoryRepo{pool: pool}
}

func (r *HistoryRepo) ListByProduct(ctx context.Context, productID uint64) ([]domain.PricePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.offer_id, m.name, h.price, h.recorded_at
		FROM price_history h
		JOIN offers o ON o.id = h.offer_id
		JOIN product_variants v ON v.id = o.product_variant_id
		JOIN merchants m ON m.id = o.merchant_id
		WHERE v.product_id = $1
		ORDER BY h.recorded_at ASC, m.name ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func (r *HistoryRepo) ListByOffer(ctx context.Context, offerID uint64) ([]domain.PricePoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT h.offer_id, m.name, h.price, h.recorded_at
		FROM price_history h
		JOIN offers o ON o.id = h.offer_id
		JOIN merchants m ON m.id = o.merchant_id
		WHERE h.offer_id = $1
		ORDER BY h.recorded_at ASC
	`, offerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPoints(rows)
}

func scanPoints(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]domain.PricePoint, error) {
	items := []domain.PricePoint{}
	for rows.Next() {
		var point domain.PricePoint
		if err := rows.Scan(&point.OfferID, &point.MerchantName, &point.Price, &point.RecordedAt); err != nil {
			return nil, err
		}
		items = append(items, point)
	}
	return items, rows.Err()
}

func (r *HistoryRepo) HasHistory(ctx context.Context, offerID uint64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM price_history WHERE offer_id = $1)`, offerID).Scan(&exists)
	return exists, err
}

func (r *HistoryRepo) Backfill(ctx context.Context, offerID uint64, points []domain.PricePoint) error {
	for _, point := range points {
		_, err := r.pool.Exec(ctx, `
			INSERT INTO price_history (offer_id, price, recorded_at)
			SELECT $1::bigint, $2::bigint, $3::timestamptz
			WHERE NOT EXISTS (
				SELECT 1 FROM price_history
				WHERE offer_id = $1::bigint AND (recorded_at AT TIME ZONE 'UTC')::date = ($3::timestamptz AT TIME ZONE 'UTC')::date
			)
		`, offerID, point.Price, point.RecordedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *HistoryRepo) AppendIfChanged(ctx context.Context, offerID uint64, price int64, at time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO price_history (offer_id, price, recorded_at)
		SELECT $1::bigint, $2::bigint, $3::timestamptz
		WHERE NOT EXISTS (
			SELECT 1 FROM price_history
			WHERE offer_id = $1::bigint AND (recorded_at AT TIME ZONE 'UTC')::date = ($3::timestamptz AT TIME ZONE 'UTC')::date
		)
		AND (
			SELECT price FROM price_history WHERE offer_id = $1::bigint ORDER BY recorded_at DESC LIMIT 1
		) IS DISTINCT FROM $2::bigint
	`, offerID, price, at)
	return err
}
