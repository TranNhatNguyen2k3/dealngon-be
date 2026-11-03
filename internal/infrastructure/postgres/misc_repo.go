package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dealngon/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CouponRepo struct{ pool *pgxpool.Pool }
type SEORepo struct{ pool *pgxpool.Pool }
type AlertRepo struct{ pool *pgxpool.Pool }
type ClickRepo struct{ pool *pgxpool.Pool }
type AdminRepo struct{ pool *pgxpool.Pool }
type SearchLogRepo struct{ pool *pgxpool.Pool }
type AuditRepo struct{ pool *pgxpool.Pool }
type EventRepo struct{ pool *pgxpool.Pool }

func NewCouponRepo(pool *pgxpool.Pool) *CouponRepo       { return &CouponRepo{pool: pool} }
func NewSEORepo(pool *pgxpool.Pool) *SEORepo             { return &SEORepo{pool: pool} }
func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo         { return &AlertRepo{pool: pool} }
func NewClickRepo(pool *pgxpool.Pool) *ClickRepo         { return &ClickRepo{pool: pool} }
func NewAdminRepo(pool *pgxpool.Pool) *AdminRepo         { return &AdminRepo{pool: pool} }
func NewSearchLogRepo(pool *pgxpool.Pool) *SearchLogRepo { return &SearchLogRepo{pool: pool} }
func NewAuditRepo(pool *pgxpool.Pool) *AuditRepo         { return &AuditRepo{pool: pool} }
func NewEventRepo(pool *pgxpool.Pool) *EventRepo         { return &EventRepo{pool: pool} }

const couponSelect = `
	SELECT c.id, c.merchant_id, c.code, c.description, c.discount_percent::float8, c.starts_at, c.ends_at, c.status, m.name, c.created_at, c.updated_at
	FROM coupons c JOIN merchants m ON m.id = c.merchant_id
`

func scanCoupon(row pgx.Row) (*domain.Coupon, error) {
	var item domain.Coupon
	err := row.Scan(&item.ID, &item.MerchantID, &item.Code, &item.Description, &item.DiscountPercent, &item.StartsAt, &item.EndsAt, &item.Status, &item.MerchantName, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CouponRepo) List(ctx context.Context, page, limit int) ([]domain.Coupon, int64, error) {
	page, limit = domain.NormalizePage(page, limit)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM coupons`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, couponSelect+` ORDER BY c.id DESC LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []domain.Coupon{}
	for rows.Next() {
		item, err := scanCoupon(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *CouponRepo) ListByMerchants(ctx context.Context, merchantIDs []uint64) ([]domain.Coupon, error) {
	if len(merchantIDs) == 0 {
		return []domain.Coupon{}, nil
	}
	rows, err := r.pool.Query(ctx, couponSelect+` WHERE c.merchant_id = ANY($1) AND c.status = 'active' ORDER BY c.id`, merchantIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []domain.Coupon{}
	for rows.Next() {
		item, err := scanCoupon(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *CouponRepo) FindByID(ctx context.Context, id uint64) (*domain.Coupon, error) {
	return scanCoupon(r.pool.QueryRow(ctx, couponSelect+` WHERE c.id = $1`, id))
}

func (r *CouponRepo) Create(ctx context.Context, coupon domain.Coupon) (*domain.Coupon, error) {
	if coupon.Status == "" {
		coupon.Status = domain.StatusActive
	}
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO coupons (merchant_id, code, description, discount_percent, starts_at, ends_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`, coupon.MerchantID, coupon.Code, coupon.Description, coupon.DiscountPercent, coupon.StartsAt, coupon.EndsAt, coupon.Status).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return r.FindByID(ctx, id)
}

func (r *CouponRepo) Update(ctx context.Context, coupon domain.Coupon) (*domain.Coupon, error) {
	tag, err := r.pool.Exec(ctx, `
		UPDATE coupons SET merchant_id = $2, code = $3, description = $4, discount_percent = $5, starts_at = $6, ends_at = $7, status = $8, updated_at = now()
		WHERE id = $1
	`, coupon.ID, coupon.MerchantID, coupon.Code, coupon.Description, coupon.DiscountPercent, coupon.StartsAt, coupon.EndsAt, coupon.Status)
	if err != nil {
		return nil, mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return nil, domain.ErrNotFound
	}
	return r.FindByID(ctx, coupon.ID)
}

func (r *CouponRepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM coupons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SEORepo) Find(ctx context.Context, entityType string, entityID uint64) (*domain.SEOMetadata, error) {
	return scanSEO(r.pool.QueryRow(ctx, `SELECT id, entity_type, entity_id, title, description, canonical, og_image, robots, updated_at FROM seo_metadata WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID))
}

func scanSEO(row pgx.Row) (*domain.SEOMetadata, error) {
	var item domain.SEOMetadata
	err := row.Scan(&item.ID, &item.EntityType, &item.EntityID, &item.Title, &item.Description, &item.Canonical, &item.OGImage, &item.Robots, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SEORepo) Upsert(ctx context.Context, meta domain.SEOMetadata) (*domain.SEOMetadata, error) {
	if meta.Robots == "" {
		meta.Robots = "index,follow"
	}
	var id uint64
	err := r.pool.QueryRow(ctx, `
		INSERT INTO seo_metadata (entity_type, entity_id, title, description, canonical, og_image, robots)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (entity_type, entity_id) DO UPDATE SET
			title = EXCLUDED.title, description = EXCLUDED.description, canonical = EXCLUDED.canonical,
			og_image = EXCLUDED.og_image, robots = EXCLUDED.robots, updated_at = now()
		RETURNING id
	`, meta.EntityType, meta.EntityID, meta.Title, meta.Description, meta.Canonical, meta.OGImage, meta.Robots).Scan(&id)
	if err != nil {
		return nil, mapError(err)
	}
	return scanSEO(r.pool.QueryRow(ctx, `SELECT id, entity_type, entity_id, title, description, canonical, og_image, robots, updated_at FROM seo_metadata WHERE id = $1`, id))
}

func (r *SEORepo) List(ctx context.Context, page, limit int) ([]domain.SEOMetadata, int64, error) {
	page, limit = domain.NormalizePage(page, limit)
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM seo_metadata`).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id, entity_type, entity_id, title, description, canonical, og_image, robots, updated_at FROM seo_metadata ORDER BY updated_at DESC LIMIT $1 OFFSET $2`, limit, (page-1)*limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []domain.SEOMetadata{}
	for rows.Next() {
		item, err := scanSEO(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, *item)
	}
	return items, total, rows.Err()
}

func (r *SEORepo) Delete(ctx context.Context, id uint64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM seo_metadata WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *AlertRepo) Create(ctx context.Context, alert domain.PriceAlert) (*domain.PriceAlert, error) {
	err := r.pool.QueryRow(ctx, `
		INSERT INTO price_alerts (product_id, email, target_price) VALUES ($1, $2, $3)
		RETURNING id, product_id, email, target_price, created_at
	`, alert.ProductID, alert.Email, alert.TargetPrice).Scan(&alert.ID, &alert.ProductID, &alert.Email, &alert.TargetPrice, &alert.CreatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &alert, nil
}

func (r *ClickRepo) Create(ctx context.Context, click domain.AffiliateClick) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO affiliate_clicks (offer_id, product_id, merchant_id, session_id, user_id, ip_hash, user_agent, referer, utm_source, utm_medium, utm_campaign)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, click.OfferID, click.ProductID, click.MerchantID, click.SessionID, click.UserID, click.IPHash, click.UserAgent, click.Referer, click.UTMSource, click.UTMMedium, click.UTMCampaign)
	return err
}

func (r *ClickRepo) SaveLink(ctx context.Context, offerID uint64, url, provider string) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO affiliate_links (offer_id, url, provider) VALUES ($1, $2, $3)`, offerID, url, provider)
	return err
}

func (r *ClickRepo) CountToday(ctx context.Context) (int64, error) {
	var n int64
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM affiliate_clicks WHERE created_at >= date_trunc('day', now())`).Scan(&n)
	return n, err
}

func (r *AdminRepo) Ensure(ctx context.Context, email, name, passwordHash string) (*domain.Admin, error) {
	var admin domain.Admin
	err := r.pool.QueryRow(ctx, `
		INSERT INTO admins (email, name, password_hash) VALUES ($1, $2, $3)
		ON CONFLICT (email) DO NOTHING
		RETURNING id, email, password_hash, name, created_at
	`, email, name, passwordHash).Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Name, &admin.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return r.FindByEmail(ctx, email)
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *AdminRepo) FindByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `SELECT id, email, password_hash, name, created_at FROM admins WHERE email = $1`, email))
}

func (r *AdminRepo) FindByID(ctx context.Context, id uint64) (*domain.Admin, error) {
	return scanAdmin(r.pool.QueryRow(ctx, `SELECT id, email, password_hash, name, created_at FROM admins WHERE id = $1`, id))
}

func scanAdmin(row pgx.Row) (*domain.Admin, error) {
	var admin domain.Admin
	err := row.Scan(&admin.ID, &admin.Email, &admin.PasswordHash, &admin.Name, &admin.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (r *SearchLogRepo) Create(ctx context.Context, query string, results int) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO search_logs (query, results_count) VALUES ($1, $2)`, query, results)
	return err
}

func (r *AuditRepo) Create(ctx context.Context, entry domain.AuditEntry) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO audit_logs (admin_id, action, entity_type, entity_id) VALUES ($1, $2, $3, $4)`, entry.AdminID, entry.Action, entry.EntityType, entry.EntityID)
	return err
}

func (r *EventRepo) Create(ctx context.Context, event domain.AnalyticsEvent) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	if event.Payload == nil {
		payload = []byte(`{}`)
	}
	_, err = r.pool.Exec(ctx, `INSERT INTO analytics_events (name, session_id, payload) VALUES ($1, $2, $3::jsonb)`, event.Name, event.SessionID, payload)
	return err
}
