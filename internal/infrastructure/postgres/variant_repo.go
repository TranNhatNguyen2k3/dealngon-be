package postgres

import (
	"context"
	"errors"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VariantRepo struct {
	pool *pgxpool.Pool
}

func NewVariantRepo(pool *pgxpool.Pool) *VariantRepo {
	return &VariantRepo{pool: pool}
}

const variantSelect = `
	SELECT id, product_id, sku, model, storage, ram, color, barcode, variant_name, created_at, updated_at
	FROM product_variants
`

func scanVariant(row pgx.Row) (*domain.Variant, error) {
	var variant domain.Variant
	err := row.Scan(
		&variant.ID, &variant.ProductID, &variant.SKU, &variant.Model, &variant.Storage,
		&variant.RAM, &variant.Color, &variant.Barcode, &variant.VariantName, &variant.CreatedAt, &variant.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &variant, nil
}

func (r *VariantRepo) ListByProduct(ctx context.Context, productID uint64) ([]domain.Variant, error) {
	rows, err := r.pool.Query(ctx, variantSelect+` WHERE product_id = $1 ORDER BY id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Variant{}
	for rows.Next() {
		variant, err := scanVariant(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *variant)
	}
	return items, rows.Err()
}

func (r *VariantRepo) FindBySKU(ctx context.Context, sku string) (*domain.Variant, error) {
	return scanVariant(r.pool.QueryRow(ctx, variantSelect+` WHERE sku = $1`, sku))
}

func (r *VariantRepo) FindByBarcode(ctx context.Context, barcode string) (*domain.Variant, error) {
	if barcode == "" {
		return nil, domain.ErrNotFound
	}
	return scanVariant(r.pool.QueryRow(ctx, variantSelect+` WHERE barcode = $1`, barcode))
}

func (r *VariantRepo) FindByBrandModel(ctx context.Context, brand, model string) (*domain.Variant, error) {
	if brand == "" || model == "" {
		return nil, domain.ErrNotFound
	}
	return scanVariant(r.pool.QueryRow(ctx, `
		SELECT v.id, v.product_id, v.sku, v.model, v.storage, v.ram, v.color, v.barcode, v.variant_name, v.created_at, v.updated_at
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		JOIN brands b ON b.id = p.brand_id
		WHERE b.slug = $1 AND lower(v.model) = lower($2)
		ORDER BY v.id
		LIMIT 1
	`, brand, model))
}

func (r *VariantRepo) FindByAttributes(ctx context.Context, brand, model, storage, ram, color string) (*domain.Variant, error) {
	if brand == "" || model == "" {
		return nil, domain.ErrNotFound
	}
	return scanVariant(r.pool.QueryRow(ctx, `
		SELECT v.id, v.product_id, v.sku, v.model, v.storage, v.ram, v.color, v.barcode, v.variant_name, v.created_at, v.updated_at
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		JOIN brands b ON b.id = p.brand_id
		WHERE b.slug = $1 AND lower(v.model) = lower($2)
			AND lower(v.storage) = lower($3) AND lower(v.ram) = lower($4) AND lower(v.color) = lower($5)
		ORDER BY v.id
		LIMIT 1
	`, brand, model, storage, ram, color))
}

func (r *VariantRepo) FindByNormalizedName(ctx context.Context, name string) (*domain.Variant, error) {
	slug := domain.Slugify(name)
	if slug == "" {
		return nil, domain.ErrNotFound
	}
	return scanVariant(r.pool.QueryRow(ctx, `
		SELECT v.id, v.product_id, v.sku, v.model, v.storage, v.ram, v.color, v.barcode, v.variant_name, v.created_at, v.updated_at
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		WHERE p.slug = $1
		ORDER BY v.id
		LIMIT 1
	`, slug))
}

func (r *VariantRepo) Upsert(ctx context.Context, variant domain.Variant) (*domain.Variant, error) {
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO product_variants (product_id, sku, model, storage, ram, color, barcode, variant_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (sku) DO UPDATE SET
			model = EXCLUDED.model,
			storage = EXCLUDED.storage,
			ram = EXCLUDED.ram,
			color = EXCLUDED.color,
			barcode = EXCLUDED.barcode,
			variant_name = EXCLUDED.variant_name,
			updated_at = now()
		RETURNING id
	`, variant.ProductID, variant.SKU, variant.Model, variant.Storage, variant.RAM, variant.Color, variant.Barcode, variant.VariantName).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return scanVariant(r.pool.QueryRow(ctx, variantSelect+` WHERE id = $1`, id))
}

func (r *VariantRepo) Update(ctx context.Context, variant domain.Variant) (*domain.Variant, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE product_variants
		SET sku = $2, model = $3, storage = $4, ram = $5, color = $6, barcode = $7, variant_name = $8, updated_at = now()
		WHERE id = $1
	`, variant.ID, variant.SKU, variant.Model, variant.Storage, variant.RAM, variant.Color, variant.Barcode, variant.VariantName)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return scanVariant(r.pool.QueryRow(ctx, variantSelect+` WHERE id = $1`, variant.ID))
}
