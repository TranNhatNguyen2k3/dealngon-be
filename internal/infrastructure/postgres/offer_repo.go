package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OfferRepo struct {
	pool *pgxpool.Pool
}

func NewOfferRepo(pool *pgxpool.Pool) *OfferRepo {
	return &OfferRepo{pool: pool}
}

const offerSelect = `
	SELECT o.id, o.product_variant_id, v.product_id, o.merchant_id, o.title, o.product_url, o.affiliate_url,
		o.price, o.original_price, o.stock, o.rating::float8, o.seller_name, o.status, o.last_checked_at,
		o.created_at, o.updated_at, m.name, m.slug, m.logo, v.variant_name, m.quality_score
	FROM offers o
	JOIN product_variants v ON v.id = o.product_variant_id
	JOIN merchants m ON m.id = o.merchant_id
`

func scanOffer(row pgx.Row) (*domain.Offer, error) {
	var offer domain.Offer
	err := row.Scan(
		&offer.ID, &offer.ProductVariantID, &offer.ProductID, &offer.MerchantID, &offer.Title,
		&offer.ProductURL, &offer.AffiliateURL, &offer.Price, &offer.OriginalPrice, &offer.Stock,
		&offer.Rating, &offer.SellerName, &offer.Status, &offer.LastCheckedAt, &offer.CreatedAt,
		&offer.UpdatedAt, &offer.MerchantName, &offer.MerchantSlug, &offer.MerchantLogo,
		&offer.VariantName, &offer.QualityScore,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *OfferRepo) FindByID(ctx context.Context, id uint64) (*domain.Offer, error) {
	return scanOffer(r.pool.QueryRow(ctx, offerSelect+` WHERE o.id = $1`, id))
}

func (r *OfferRepo) ListByProduct(ctx context.Context, productID uint64) ([]domain.Offer, error) {
	rows, err := r.pool.Query(ctx, offerSelect+` WHERE v.product_id = $1 AND o.status = 'active' ORDER BY o.price ASC, o.id ASC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectOffers(rows)
}

func (r *OfferRepo) List(ctx context.Context, page, limit int) ([]domain.Offer, int64, error) {
	page, limit = domain.NormalizePage(page, limit)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM offers`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, offerSelect+` ORDER BY o.updated_at DESC LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := collectOffers(rows)
	return items, total, err
}

func collectOffers(rows pgx.Rows) ([]domain.Offer, error) {
	items := []domain.Offer{}
	for rows.Next() {
		offer, err := scanOffer(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *offer)
	}
	return items, rows.Err()
}

func (r *OfferRepo) Upsert(ctx context.Context, offer domain.Offer) (*domain.Offer, error) {
	if offer.Status == "" {
		offer.Status = domain.StatusActive
	}
	if offer.LastCheckedAt == nil {
		now := time.Now().UTC()
		offer.LastCheckedAt = &now
	}
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO offers (
			product_variant_id, merchant_id, title, product_url, price, original_price,
			stock, rating, seller_name, status, last_checked_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (product_variant_id, merchant_id) DO UPDATE SET
			title = EXCLUDED.title,
			product_url = EXCLUDED.product_url,
			price = EXCLUDED.price,
			original_price = EXCLUDED.original_price,
			stock = EXCLUDED.stock,
			rating = EXCLUDED.rating,
			seller_name = EXCLUDED.seller_name,
			status = EXCLUDED.status,
			last_checked_at = EXCLUDED.last_checked_at,
			updated_at = now()
		RETURNING id
	`, offer.ProductVariantID, offer.MerchantID, offer.Title, offer.ProductURL, offer.Price, offer.OriginalPrice,
		offer.Stock, offer.Rating, offer.SellerName, offer.Status, offer.LastCheckedAt).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *OfferRepo) Update(ctx context.Context, offer domain.Offer) (*domain.Offer, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE offers SET
			product_variant_id = $2, merchant_id = $3, title = $4, product_url = $5,
			price = $6, original_price = $7, stock = $8, rating = $9, seller_name = $10,
			status = $11, updated_at = now()
		WHERE id = $1
	`, offer.ID, offer.ProductVariantID, offer.MerchantID, offer.Title, offer.ProductURL,
		offer.Price, offer.OriginalPrice, offer.Stock, offer.Rating, offer.SellerName, offer.Status)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, offer.ID)
}

func (r *OfferRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM offers WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *OfferRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM offers`).Scan(&n)
	return n, err
}

func (r *OfferRepo) SetAffiliateURL(ctx context.Context, offerID uint64, affiliateURL string) error {
	tag, err := r.pool.Exec(ctx, `UPDATE offers SET affiliate_url = $2, updated_at = now() WHERE id = $1`, offerID, affiliateURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
