package domain

import (
	"math"
	"time"
)

func HistoricalPrice(current int64, daysAgo int, allowDip bool) int64 {
	if current <= 0 {
		return 0
	}
	if daysAgo <= 0 {
		return current
	}
	mult := 1.0 + float64(daysAgo)*0.005
	if allowDip && daysAgo == 14 {
		mult = 0.96
	}
	rounded := int64(math.Round(float64(current)*mult/1000)) * 1000
	if rounded < 1000 {
		return current
	}
	return rounded
}

func AllowPriceDip(sku string) bool {
	if sku == "" {
		return false
	}
	return sku[len(sku)-1]%2 == 0
}

func SummarizeHistory(points []PricePoint, now time.Time) (avg30, min30, histMin int64) {
	if len(points) == 0 {
		return 0, 0, 0
	}
	histMin = points[0].Price
	cutoff := now.Add(-30 * 24 * time.Hour)
	var sum int64
	var n int64
	for _, point := range points {
		if point.Price <= 0 {
			continue
		}
		if point.Price < histMin {
			histMin = point.Price
		}
		if !point.RecordedAt.Before(cutoff) {
			sum += point.Price
			n++
			if min30 == 0 || point.Price < min30 {
				min30 = point.Price
			}
		}
	}
	if n > 0 {
		avg30 = sum / n
	}
	return avg30, min30, histMin
}

func SummarizeOffers(offers []Offer) (lowest, average, savings *int64, last *time.Time) {
	var sum int64
	var n int64
	for _, offer := range offers {
		if offer.Status != "" && offer.Status != StatusActive {
			continue
		}
		if offer.Price <= 0 {
			continue
		}
		if lowest == nil || offer.Price < *lowest {
			price := offer.Price
			lowest = &price
		}
		sum += offer.Price
		n++
		if offer.LastCheckedAt != nil && (last == nil || offer.LastCheckedAt.After(*last)) {
			checked := *offer.LastCheckedAt
			last = &checked
		}
	}
	if n == 0 {
		return nil, nil, nil, last
	}
	avg := sum / n
	average = &avg
	if lowest != nil && avg > *lowest {
		saved := avg - *lowest
		savings = &saved
	}
	return lowest, average, savings, last
}

func BestOffer(offers []Offer) *Offer {
	var best *Offer
	for i := range offers {
		offer := &offers[i]
		if offer.Status != "" && offer.Status != StatusActive {
			continue
		}
		if offer.Price <= 0 {
			continue
		}
		if best == nil || offer.Price < best.Price {
			best = offer
		}
	}
	return best
}

func NormalizePage(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func NormalizeSort(sort string) string {
	switch sort {
	case "price_asc", "price_desc", "discount", "deal", "popular", "relevance":
		return sort
	default:
		return "relevance"
	}
}
