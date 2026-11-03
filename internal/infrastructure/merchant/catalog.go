package merchant

import (
	"context"
	"fmt"
	"net/url"

	"github.com/dealngon/backend/internal/domain"
)

type Info struct {
	Name         string
	Slug         string
	Website      string
	QualityScore int
	SearchURL    string
}

func Directory() []Info {
	return []Info{
		{Name: "Shopee", Slug: "shopee", Website: "https://shopee.vn", QualityScore: 4, SearchURL: "https://shopee.vn/search?keyword="},
		{Name: "Lazada", Slug: "lazada", Website: "https://www.lazada.vn", QualityScore: 4, SearchURL: "https://www.lazada.vn/catalog/?q="},
		{Name: "TikTok Shop", Slug: "tiktok-shop", Website: "https://shop.tiktok.com", QualityScore: 3, SearchURL: "https://shop.tiktok.com/vn/search?q="},
		{Name: "CellphoneS", Slug: "cellphones", Website: "https://cellphones.com.vn", QualityScore: 5, SearchURL: "https://cellphones.com.vn/catalogsearch/result?q="},
		{Name: "Điện Máy Xanh", Slug: "dien-may-xanh", Website: "https://www.dienmayxanh.com", QualityScore: 5, SearchURL: "https://www.dienmayxanh.com/tim-kiem?key="},
	}
}

func FindInfo(slug string) (Info, bool) {
	for _, info := range Directory() {
		if info.Slug == slug {
			return info, true
		}
	}
	return Info{}, false
}

type priceQuote struct {
	Merchant string
	Price    int64
	Original *int64
}

type catalogItem struct {
	Name           string
	Brand          string
	Category       string
	CategoryName   string
	SKU            string
	Model          string
	Storage        string
	RAM            string
	Color          string
	Barcode        string
	Description    string
	VariantName    string
	Specifications map[string]string
	Quotes         []priceQuote
}

func money(value int64) *int64 { return &value }

func items() []catalogItem {
	return []catalogItem{
		{
			Name: "iPhone 16 128GB", Brand: "Apple", Category: "dien-thoai", CategoryName: "Điện thoại",
			SKU: "IP16-128-BLK", Model: "iPhone 16", Storage: "128GB", Color: "Đen", Barcode: "194253715001",
			VariantName:    "128GB Đen",
			Description:    "Điện thoại Apple iPhone 16 bộ nhớ 128GB, màu đen.",
			Specifications: map[string]string{"Màn hình": "6.1 inch", "Chip": "A18", "Bộ nhớ": "128GB", "Màu": "Đen"},
			Quotes: []priceQuote{
				{"shopee", 22190000, money(24990000)},
				{"lazada", 22790000, money(24990000)},
				{"tiktok-shop", 22490000, money(24990000)},
				{"cellphones", 22990000, money(24990000)},
				{"dien-may-xanh", 23490000, money(24990000)},
			},
		},
		{
			Name: "Samsung Galaxy S24 256GB", Brand: "Samsung", Category: "dien-thoai", CategoryName: "Điện thoại",
			SKU: "SM-S921-256", Model: "Galaxy S24", Storage: "256GB", Color: "Xám", Barcode: "8806095490016",
			VariantName:    "256GB Xám",
			Description:    "Điện thoại Samsung Galaxy S24 bộ nhớ 256GB.",
			Specifications: map[string]string{"Màn hình": "6.2 inch", "RAM": "8GB", "Bộ nhớ": "256GB", "Màu": "Xám"},
			Quotes: []priceQuote{
				{"shopee", 16490000, money(18990000)},
				{"lazada", 16990000, money(18990000)},
				{"tiktok-shop", 16790000, money(18990000)},
				{"cellphones", 17290000, money(18990000)},
				{"dien-may-xanh", 17490000, money(18990000)},
			},
		},
		{
			Name: "MacBook Air 13 inch M3 256GB", Brand: "Apple", Category: "laptop", CategoryName: "Laptop",
			SKU: "MBA-M3-256", Model: "MacBook Air 13 M3", Storage: "256GB", RAM: "16GB", Color: "Bạc",
			VariantName:    "16GB/256GB Bạc",
			Description:    "Laptop Apple MacBook Air 13 inch chip M3, RAM 16GB, SSD 256GB.",
			Specifications: map[string]string{"Màn hình": "13.6 inch", "Chip": "Apple M3", "RAM": "16GB", "SSD": "256GB"},
			Quotes: []priceQuote{
				{"cellphones", 26990000, money(27990000)},
				{"dien-may-xanh", 27490000, money(27990000)},
				{"shopee", 27290000, nil},
				{"lazada", 27690000, nil},
			},
		},
		{
			Name: "iPad Air 11 inch M2 128GB", Brand: "Apple", Category: "tablet", CategoryName: "Tablet",
			SKU: "IPAD-AIR11-128", Model: "iPad Air 11 M2", Storage: "128GB", Color: "Xanh",
			VariantName:    "128GB Xanh",
			Description:    "Máy tính bảng Apple iPad Air 11 inch chip M2, bộ nhớ 128GB.",
			Specifications: map[string]string{"Màn hình": "11 inch", "Chip": "Apple M2", "Bộ nhớ": "128GB"},
			Quotes: []priceQuote{
				{"shopee", 16490000, money(17990000)},
				{"lazada", 16790000, money(17990000)},
				{"cellphones", 16990000, money(17990000)},
				{"dien-may-xanh", 17190000, money(17990000)},
			},
		},
		{
			Name: "Sony WH-1000XM5", Brand: "Sony", Category: "tai-nghe", CategoryName: "Tai nghe",
			SKU: "WH1000XM5-BLK", Model: "WH-1000XM5", Color: "Đen", Barcode: "4548736142475",
			VariantName:    "Đen",
			Description:    "Tai nghe chống ồn Sony WH-1000XM5 màu đen.",
			Specifications: map[string]string{"Kết nối": "Bluetooth", "Chống ồn": "Có", "Màu": "Đen"},
			Quotes: []priceQuote{
				{"tiktok-shop", 6890000, money(8490000)},
				{"shopee", 6990000, money(8490000)},
				{"lazada", 7290000, money(8490000)},
				{"cellphones", 7490000, money(8490000)},
				{"dien-may-xanh", 7690000, money(8490000)},
			},
		},
		{
			Name: "Samsung Crystal UHD 55 inch", Brand: "Samsung", Category: "tv", CategoryName: "TV",
			SKU: "UA55CU8000", Model: "UA55CU8000",
			VariantName:    "55 inch",
			Description:    "Smart TV Samsung Crystal UHD 55 inch.",
			Specifications: map[string]string{"Kích thước": "55 inch", "Độ phân giải": "4K UHD", "Hệ điều hành": "Tizen"},
			Quotes: []priceQuote{
				{"dien-may-xanh", 10990000, money(12990000)},
				{"shopee", 11290000, money(12990000)},
				{"cellphones", 11490000, money(12990000)},
				{"lazada", 11690000, money(12990000)},
			},
		},
		{
			Name: "Anker PowerCore 20000mAh", Brand: "Anker", Category: "phu-kien", CategoryName: "Phụ kiện",
			SKU: "A1363", Model: "PowerCore 20000", Barcode: "194644116548",
			VariantName:    "20000mAh",
			Description:    "Pin sạc dự phòng Anker PowerCore dung lượng 20000mAh.",
			Specifications: map[string]string{"Dung lượng": "20000mAh", "Cổng": "USB-C"},
			Quotes: []priceQuote{
				{"shopee", 890000, money(1290000)},
				{"tiktok-shop", 950000, money(1290000)},
				{"lazada", 990000, money(1290000)},
			},
		},
		{
			Name: "Xiaomi Redmi Buds 5", Brand: "Xiaomi", Category: "tai-nghe", CategoryName: "Tai nghe",
			SKU: "REDMI-BUDS5", Model: "Redmi Buds 5", Color: "Trắng",
			VariantName:    "Trắng",
			Description:    "Tai nghe true wireless Xiaomi Redmi Buds 5.",
			Specifications: map[string]string{"Kết nối": "Bluetooth", "Màu": "Trắng"},
			Quotes: []priceQuote{
				{"tiktok-shop", 450000, money(690000)},
				{"shopee", 490000, money(690000)},
				{"lazada", 520000, money(690000)},
			},
		},
	}
}

func CategoryNames() map[string]string {
	names := map[string]string{}
	for _, item := range items() {
		names[item.Category] = item.CategoryName
	}
	return names
}

func ProductsFor(slug string) []domain.ExternalProduct {
	info, ok := FindInfo(slug)
	if !ok {
		return nil
	}
	var result []domain.ExternalProduct
	for _, item := range items() {
		for _, quote := range item.Quotes {
			if quote.Merchant != slug {
				continue
			}
			result = append(result, toExternal(item, info, quote))
		}
	}
	return result
}

func OffersFor(slug string) []domain.ExternalOffer {
	products := ProductsFor(slug)
	offers := make([]domain.ExternalOffer, 0, len(products))
	for _, product := range products {
		offers = append(offers, domain.ExternalOffer{
			MerchantSlug:  product.MerchantSlug,
			SKU:           product.SKU,
			Barcode:       product.Barcode,
			Title:         product.Name,
			ProductURL:    product.ProductURL,
			Price:         product.Price,
			OriginalPrice: product.OriginalPrice,
			Stock:         product.Stock,
			Rating:        product.Rating,
			SellerName:    product.SellerName,
		})
	}
	return offers
}

func toExternal(item catalogItem, info Info, quote priceQuote) domain.ExternalProduct {
	return domain.ExternalProduct{
		MerchantSlug:   info.Slug,
		Name:           item.Name,
		Brand:          item.Brand,
		Category:       item.Category,
		SKU:            item.SKU,
		Model:          item.Model,
		Storage:        item.Storage,
		RAM:            item.RAM,
		Color:          item.Color,
		Barcode:        item.Barcode,
		Description:    item.Description,
		Specifications: item.Specifications,
		VariantName:    item.VariantName,
		ProductURL:     info.SearchURL + url.QueryEscape(item.Name),
		Price:          quote.Price,
		OriginalPrice:  quote.Original,
		SellerName:     info.Name,
	}
}

type Provider struct {
	info Info
}

func NewProvider(slug string) (Provider, error) {
	info, ok := FindInfo(slug)
	if !ok {
		return Provider{}, fmt.Errorf("unknown merchant %s", slug)
	}
	return Provider{info: info}, nil
}

func AllProviders() []domain.MerchantProvider {
	directory := Directory()
	providers := make([]domain.MerchantProvider, 0, len(directory))
	for _, info := range directory {
		providers = append(providers, Provider{info: info})
	}
	return providers
}

func (p Provider) Slug() string { return p.info.Slug }

func (p Provider) FetchProducts(ctx context.Context) ([]domain.ExternalProduct, error) {
	return ProductsFor(p.info.Slug), nil
}

func (p Provider) FetchOffers(ctx context.Context) ([]domain.ExternalOffer, error) {
	return OffersFor(p.info.Slug), nil
}
