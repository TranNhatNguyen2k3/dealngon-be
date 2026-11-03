package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

const productSelect = `
	SELECT p.id, p.brand_id, p.category_id, p.name, p.slug, p.description, p.image,
		p.specifications, p.status, p.created_at, p.updated_at,
		b.name, b.slug, c.name, c.slug
	FROM products p
	JOIN brands b ON b.id = p.brand_id
	JOIN categories c ON c.id = p.category_id
`

func scanProduct(row pgx.Row) (*domain.Product, error) {
	var product domain.Product
	var specs []byte
	err := row.Scan(
		&product.ID, &product.BrandID, &product.CategoryID, &product.Name, &product.Slug,
		&product.Description, &product.Image, &specs, &product.Status, &product.CreatedAt, &product.UpdatedAt,
		&product.BrandName, &product.BrandSlug, &product.CategoryName, &product.CategorySlug,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	product.Specifications = map[string]string{}
	if len(specs) > 0 {
		_ = json.Unmarshal(specs, &product.Specifications)
	}
	return &product, nil
}

func (r *ProductRepo) FindByID(ctx context.Context, id uint64) (*domain.Product, error) {
	return scanProduct(r.pool.QueryRow(ctx, productSelect+" WHERE p.id = $1", id))
}

func (r *ProductRepo) FindBySlug(ctx context.Context, slug string) (*domain.Product, error) {
	return scanProduct(r.pool.QueryRow(ctx, productSelect+" WHERE p.slug = $1", slug))
}

func (r *ProductRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&n)
	return n, err
}

func (r *ProductRepo) ListIDs(ctx context.Context) ([]uint64, error) {
	rows, err := r.pool.Query(ctx, `SELECT id FROM products WHERE status = 'active' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *ProductRepo) List(ctx context.Context, page, limit int) ([]domain.Product, int64, error) {
	page, limit = domain.NormalizePage(page, limit)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM products`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, productSelect+` ORDER BY p.updated_at DESC LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []domain.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *product)
	}
	if items == nil {
		items = []domain.Product{}
	}
	return items, total, rows.Err()
}

func (r *ProductRepo) Create(ctx context.Context, product domain.Product) (*domain.Product, error) {
	specs, err := json.Marshal(product.Specifications)
	if err != nil {
		return nil, err
	}
	if product.Status == "" {
		product.Status = domain.StatusActive
	}
	var id uint64
	err = r.pool.QueryRow(ctx, `
		INSERT INTO products (brand_id, category_id, name, slug, description, image, specifications, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
		RETURNING id
	`, product.BrandID, product.CategoryID, product.Name, product.Slug, product.Description, product.Image, specs, product.Status).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *ProductRepo) Update(ctx context.Context, product domain.Product) (*domain.Product, error) {
	specs, err := json.Marshal(product.Specifications)
	if err != nil {
		return nil, err
	}
	tag, err := r.pool.Exec(ctx, `
		UPDATE products
		SET brand_id = $2, category_id = $3, name = $4, slug = $5, description = $6,
			image = $7, specifications = $8::jsonb, status = $9, updated_at = now()
		WHERE id = $1
	`, product.ID, product.BrandID, product.CategoryID, product.Name, product.Slug, product.Description, product.Image, specs, product.Status)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, product.ID)
}

func (r *ProductRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *ProductRepo) ListSitemap(ctx context.Context) ([]domain.SitemapEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT slug, updated_at FROM products WHERE status = 'active' ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.SitemapEntry
	for rows.Next() {
		var slug string
		var updated time.Time
		if err := rows.Scan(&slug, &updated); err != nil {
			return nil, err
		}
		items = append(items, domain.SitemapEntry{Path: "/san-pham/" + slug, UpdatedAt: updated})
		items = append(items, domain.SitemapEntry{Path: "/gia/" + slug, UpdatedAt: updated})
	}
	return items, rows.Err()
}

func (r *ProductRepo) Search(ctx context.Context, query domain.SearchQuery) ([]domain.ProductCard, int64, error) {
	page, limit := domain.NormalizePage(query.Page, query.Limit)
	sort := domain.NormalizeSort(query.Sort)
	where := []string{"p.status = 'active'"}
	args := make([]any, 0, 8)

	add := func(expr string, value any) {
		args = append(args, value)
		where = append(where, fmt.Sprintf(expr, len(args)))
	}
	if text := strings.TrimSpace(query.Query); text != "" {
		args = append(args, "%"+text+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR b.name ILIKE $%d OR COALESCE(v.model, '') ILIKE $%d OR COALESCE(v.sku, '') ILIKE $%d)", n, n, n, n))
	}
	if query.CategorySlug != "" {
		add("c.slug = $%d", query.CategorySlug)
	}
	if query.BrandSlug != "" {
		add("b.slug = $%d", query.BrandSlug)
	}
	if query.MerchantSlug != "" {
		add("m.slug = $%d", query.MerchantSlug)
	}

	having := make([]string, 0, 3)
	addHaving := func(expr string, value any) {
		args = append(args, value)
		having = append(having, fmt.Sprintf(expr, len(args)))
	}
	if query.MinPrice != nil {
		addHaving("MIN(o.price) FILTER (WHERE o.status = 'active') >= $%d", *query.MinPrice)
	}
	if query.MaxPrice != nil {
		addHaving("MIN(o.price) FILTER (WHERE o.status = 'active') <= $%d", *query.MaxPrice)
	}
	if query.MinDiscount != nil {
		addHaving(`MAX(CASE WHEN o.status = 'active' AND o.original_price IS NOT NULL AND o.original_price > o.price
			THEN (o.original_price - o.price)::numeric * 100 / o.original_price ELSE 0 END) >= $%d`, *query.MinDiscount)
	}

	from := `
		FROM products p
		JOIN brands b ON b.id = p.brand_id
		JOIN categories c ON c.id = p.category_id
		LEFT JOIN product_variants v ON v.product_id = p.id
		LEFT JOIN offers o ON o.product_variant_id = v.id
		LEFT JOIN merchants m ON m.id = o.merchant_id
		LEFT JOIN deals d ON d.product_id = p.id
	`
	whereSQL := "WHERE " + strings.Join(where, " AND ")
	groupSQL := "GROUP BY p.id, b.id, c.id, d.id"
	havingSQL := ""
	if len(having) > 0 {
		havingSQL = "HAVING " + strings.Join(having, " AND ")
	}

	var total int64
	countSQL := "SELECT COUNT(*) FROM (SELECT p.id " + from + whereSQL + " " + groupSQL + " " + havingSQL + ") counted"
	if err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	order := "p.name ASC"
	switch sort {
	case "price_asc":
		order = "lowest_price ASC NULLS LAST, p.name ASC"
	case "price_desc":
		order = "lowest_price DESC NULLS LAST, p.name ASC"
	case "discount":
		order = "discount_percent DESC NULLS LAST, p.name ASC"
	case "deal":
		order = "d.score DESC NULLS LAST, lowest_price ASC NULLS LAST"
	case "popular":
		order = "offer_count DESC, p.name ASC"
	default:
		if text := strings.TrimSpace(query.Query); text != "" {
			args = append(args, strings.ToLower(text))
			n := len(args)
			order = fmt.Sprintf("CASE WHEN lower(p.name) = $%d THEN 0 WHEN lower(p.name) LIKE $%d || '%%' THEN 1 ELSE 2 END, p.name ASC", n, n)
		}
	}
	args = append(args, limit, (page-1)*limit)
	limitPos := len(args) - 1
	offsetPos := len(args)
	listSQL := fmt.Sprintf(`
		SELECT p.id, p.name, p.slug, p.image,
			b.name, b.slug, c.name, c.slug,
			MIN(o.price) FILTER (WHERE o.status = 'active') AS lowest_price,
			(ARRAY_AGG(o.original_price ORDER BY o.price ASC) FILTER (WHERE o.status = 'active'))[1] AS original_price,
			d.score, COALESCE(d.label, ''),
			COUNT(o.id) FILTER (WHERE o.status = 'active') AS offer_count,
			MAX(o.last_checked_at) AS last_checked_at
		%s %s %s %s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, from, whereSQL, groupSQL, havingSQL, order, limitPos, offsetPos)

	rows, err := r.pool.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []domain.ProductCard{}
	for rows.Next() {
		var card domain.ProductCard
		var offerCount int64
		if err := rows.Scan(
			&card.ID, &card.Name, &card.Slug, &card.Image,
			&card.BrandName, &card.BrandSlug, &card.CategoryName, &card.CategorySlug,
			&card.LowestPrice, &card.OriginalPrice, &card.DealScore, &card.DealLabel,
			&offerCount, &card.LastCheckedAt,
		); err != nil {
			return nil, 0, err
		}
		card.OfferCount = int(offerCount)
		if card.LowestPrice != nil {
			_, card.DiscountPercent = domain.DiscountFromPrices(*card.LowestPrice, card.OriginalPrice)
		}
		items = append(items, card)
	}
	return items, total, rows.Err()
}
