package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BrandRepo struct{ pool *pgxpool.Pool }
type CategoryRepo struct{ pool *pgxpool.Pool }
type MerchantRepo struct{ pool *pgxpool.Pool }

func NewBrandRepo(pool *pgxpool.Pool) *BrandRepo       { return &BrandRepo{pool: pool} }
func NewCategoryRepo(pool *pgxpool.Pool) *CategoryRepo { return &CategoryRepo{pool: pool} }
func NewMerchantRepo(pool *pgxpool.Pool) *MerchantRepo { return &MerchantRepo{pool: pool} }

func (r *BrandRepo) List(ctx context.Context) ([]domain.Brand, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug, logo, created_at, updated_at FROM brands ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Brand{}
	for rows.Next() {
		var item domain.Brand
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Logo, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *BrandRepo) FindBySlug(ctx context.Context, slug string) (*domain.Brand, error) {
	return scanBrand(r.pool.QueryRow(ctx, `SELECT id, name, slug, logo, created_at, updated_at FROM brands WHERE slug = $1`, slug))
}

func (r *BrandRepo) FindByID(ctx context.Context, id uint64) (*domain.Brand, error) {
	return scanBrand(r.pool.QueryRow(ctx, `SELECT id, name, slug, logo, created_at, updated_at FROM brands WHERE id = $1`, id))
}

func scanBrand(row pgx.Row) (*domain.Brand, error) {
	var item domain.Brand
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.Logo, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *BrandRepo) Upsert(ctx context.Context, brand domain.Brand) (*domain.Brand, error) {
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO brands (name, slug, logo) VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
		RETURNING id
	`, brand.Name, brand.Slug, brand.Logo).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *BrandRepo) Update(ctx context.Context, brand domain.Brand) (*domain.Brand, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE brands SET name = $2, slug = $3, logo = $4, updated_at = now() WHERE id = $1`, brand.ID, brand.Name, brand.Slug, brand.Logo)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, brand.ID)
}

func (r *BrandRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM brands WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BrandRepo) ListSitemap(ctx context.Context) ([]domain.SitemapEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT slug, updated_at FROM brands ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSitemap(rows, "/thuong-hieu/")
}

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, slug, parent_id, created_at, updated_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Category{}
	for rows.Next() {
		item, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *CategoryRepo) FindBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	return scanCategory(r.pool.QueryRow(ctx, `SELECT id, name, slug, parent_id, created_at, updated_at FROM categories WHERE slug = $1`, slug))
}

func (r *CategoryRepo) FindByID(ctx context.Context, id uint64) (*domain.Category, error) {
	return scanCategory(r.pool.QueryRow(ctx, `SELECT id, name, slug, parent_id, created_at, updated_at FROM categories WHERE id = $1`, id))
}

func scanCategory(row pgx.Row) (*domain.Category, error) {
	var item domain.Category
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.ParentID, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CategoryRepo) Upsert(ctx context.Context, category domain.Category) (*domain.Category, error) {
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name, slug, parent_id) VALUES ($1, $2, $3)
		ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, updated_at = now()
		RETURNING id
	`, category.Name, category.Slug, category.ParentID).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *CategoryRepo) Update(ctx context.Context, category domain.Category) (*domain.Category, error) {
	tag, err := r.pool.Exec(ctx, `UPDATE categories SET name = $2, slug = $3, parent_id = $4, updated_at = now() WHERE id = $1`, category.ID, category.Name, category.Slug, category.ParentID)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, category.ID)
}

func (r *CategoryRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *CategoryRepo) ListSitemap(ctx context.Context) ([]domain.SitemapEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT slug, updated_at FROM categories ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSitemap(rows, "/danh-muc/")
}

func scanSitemap(rows pgx.Rows, prefix string) ([]domain.SitemapEntry, error) {
	items := []domain.SitemapEntry{}
	for rows.Next() {
		var slug string
		var updated time.Time
		if err := rows.Scan(&slug, &updated); err != nil {
			return nil, err
		}
		items = append(items, domain.SitemapEntry{Path: prefix + slug, UpdatedAt: updated})
	}
	return items, rows.Err()
}

func (r *MerchantRepo) List(ctx context.Context) ([]domain.Merchant, error) {
	rows, err := r.pool.Query(ctx, merchantSelect+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Merchant{}
	for rows.Next() {
		item, err := scanMerchant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

const merchantSelect = `SELECT id, name, slug, logo, website, affiliate_enabled, status, quality_score, created_at, updated_at FROM merchants`

func (r *MerchantRepo) FindBySlug(ctx context.Context, slug string) (*domain.Merchant, error) {
	return scanMerchant(r.pool.QueryRow(ctx, merchantSelect+` WHERE slug = $1`, slug))
}

func (r *MerchantRepo) FindByID(ctx context.Context, id uint64) (*domain.Merchant, error) {
	return scanMerchant(r.pool.QueryRow(ctx, merchantSelect+` WHERE id = $1`, id))
}

func scanMerchant(row pgx.Row) (*domain.Merchant, error) {
	var item domain.Merchant
	err := row.Scan(&item.ID, &item.Name, &item.Slug, &item.Logo, &item.Website, &item.AffiliateEnabled, &item.Status, &item.QualityScore, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MerchantRepo) Upsert(ctx context.Context, merchant domain.Merchant) (*domain.Merchant, error) {
	if merchant.Status == "" {
		merchant.Status = domain.StatusActive
	}
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO merchants (name, slug, logo, website, affiliate_enabled, status, quality_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (slug) DO UPDATE SET
			name = EXCLUDED.name,
			website = EXCLUDED.website,
			affiliate_enabled = EXCLUDED.affiliate_enabled,
			quality_score = EXCLUDED.quality_score,
			updated_at = now()
		RETURNING id
	`, merchant.Name, merchant.Slug, merchant.Logo, merchant.Website, merchant.AffiliateEnabled, merchant.Status, merchant.QualityScore).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *MerchantRepo) Update(ctx context.Context, merchant domain.Merchant) (*domain.Merchant, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE merchants SET name = $2, slug = $3, logo = $4, website = $5, affiliate_enabled = $6, status = $7, quality_score = $8, updated_at = now()
		WHERE id = $1
	`, merchant.ID, merchant.Name, merchant.Slug, merchant.Logo, merchant.Website, merchant.AffiliateEnabled, merchant.Status, merchant.QualityScore)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, merchant.ID)
}

func (r *MerchantRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM merchants WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
