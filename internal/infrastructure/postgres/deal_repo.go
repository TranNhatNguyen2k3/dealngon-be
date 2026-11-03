package postgres

import (
	"context"
	"errors"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DealRepo struct{ pool *pgxpool.Pool }

func NewDealRepo(pool *pgxpool.Pool) *DealRepo { return &DealRepo{pool: pool} }

const dealSelect = `
	SELECT d.id, d.product_id, d.offer_id, d.slug, d.score, d.label, d.discount_amount,
		d.discount_percent::float8, d.lowest_price, d.average_price, d.status, d.detected_at,
		p.name, p.slug, p.image, b.name, m.name, c.slug
	FROM deals d
	JOIN products p ON p.id = d.product_id
	JOIN brands b ON b.id = p.brand_id
	JOIN categories c ON c.id = p.category_id
	JOIN offers o ON o.id = d.offer_id
	JOIN merchants m ON m.id = o.merchant_id
`

func scanDeal(row pgx.Row) (*domain.Deal, error) {
	var deal domain.Deal
	err := row.Scan(
		&deal.ID, &deal.ProductID, &deal.OfferID, &deal.Slug, &deal.Score, &deal.Label,
		&deal.DiscountAmount, &deal.DiscountPercent, &deal.LowestPrice, &deal.AveragePrice,
		&deal.Status, &deal.DetectedAt, &deal.ProductName, &deal.ProductSlug, &deal.ProductImage,
		&deal.BrandName, &deal.MerchantName, &deal.CategorySlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &deal, nil
}

func (r *DealRepo) List(ctx context.Context, query domain.SearchQuery) ([]domain.Deal, int64, error) {
	page, limit := domain.NormalizePage(query.Page, query.Limit)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deals WHERE status = 'active'`).Scan(&total); err != nil {
		return nil, 0, err
	}
	order := "d.score DESC, d.lowest_price ASC"
	if query.Sort == "price_asc" {
		order = "d.lowest_price ASC"
	}
	rows, err := r.pool.Query(ctx, dealSelect+` WHERE d.status = 'active' ORDER BY `+order+` LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []domain.Deal{}
	for rows.Next() {
		deal, err := scanDeal(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *deal)
	}
	return items, total, rows.Err()
}

func (r *DealRepo) FindByID(ctx context.Context, id uint64) (*domain.Deal, error) {
	return scanDeal(r.pool.QueryRow(ctx, dealSelect+` WHERE d.id = $1`, id))
}

func (r *DealRepo) FindBySlug(ctx context.Context, slug string) (*domain.Deal, error) {
	return scanDeal(r.pool.QueryRow(ctx, dealSelect+` WHERE d.slug = $1`, slug))
}

func (r *DealRepo) Upsert(ctx context.Context, deal domain.Deal) (*domain.Deal, error) {
	if deal.Status == "" {
		deal.Status = domain.StatusActive
	}
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO deals (product_id, offer_id, slug, score, label, discount_amount, discount_percent, lowest_price, average_price, status, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (product_id) DO UPDATE SET
			offer_id = EXCLUDED.offer_id,
			slug = EXCLUDED.slug,
			score = EXCLUDED.score,
			label = EXCLUDED.label,
			discount_amount = EXCLUDED.discount_amount,
			discount_percent = EXCLUDED.discount_percent,
			lowest_price = EXCLUDED.lowest_price,
			average_price = EXCLUDED.average_price,
			status = EXCLUDED.status,
			detected_at = EXCLUDED.detected_at
		RETURNING id
	`, deal.ProductID, deal.OfferID, deal.Slug, deal.Score, deal.Label, deal.DiscountAmount, deal.DiscountPercent, deal.LowestPrice, deal.AveragePrice, deal.Status, deal.DetectedAt).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *DealRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM deals WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DealRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deals`).Scan(&n)
	return n, err
}

func (r *DealRepo) UpdateStatus(ctx context.Context, id uint64, status string) (*domain.Deal, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE deals SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, id)
}
