package domain

import "math"

type DealScoreInput struct {
	CurrentPrice    int64
	OriginalPrice   *int64
	Avg30           int64
	Min30           int64
	HistoricalMin   int64
	MerchantQuality int
}

type DealScoreResult struct {
	Score           int
	Label           string
	DiscountAmount  *int64
	DiscountPercent *float64
}

func CalculateDealScore(in DealScoreInput) DealScoreResult {
	score := 0.0
	var discountAmount *int64
	var discountPercent *float64

	if in.Avg30 > 0 && in.CurrentPrice > 0 && in.CurrentPrice < in.Avg30 {
		ratio := float64(in.Avg30-in.CurrentPrice) / float64(in.Avg30)
		if ratio > 1 {
			ratio = 1
		}
		score += ratio * 35
	}

	if in.Min30 > 0 && in.CurrentPrice > 0 {
		if in.CurrentPrice <= in.Min30 {
			score += 25
		} else {
			gap := float64(in.CurrentPrice-in.Min30) / float64(in.Min30)
			if gap < 0.05 {
				score += (1 - gap/0.05) * 15
			}
		}
	}

	if in.OriginalPrice != nil && *in.OriginalPrice > in.CurrentPrice && in.CurrentPrice > 0 {
		amount := *in.OriginalPrice - in.CurrentPrice
		percent := float64(amount) / float64(*in.OriginalPrice) * 100
		discountAmount = &amount
		discountPercent = &percent
		ratio := percent / 100
		if ratio > 1 {
			ratio = 1
		}
		score += ratio * 20
	}

	if in.HistoricalMin > 0 && in.CurrentPrice > 0 && in.CurrentPrice <= in.HistoricalMin {
		score += 15
	}

	quality := in.MerchantQuality
	if quality < 0 {
		quality = 0
	}
	if quality > 5 {
		quality = 5
	}
	score += float64(quality)

	if score > 100 {
		score = 100
	}
	rounded := int(math.Round(score))
	if rounded > 100 {
		rounded = 100
	}
	if rounded < 0 {
		rounded = 0
	}
	return DealScoreResult{
		Score:           rounded,
		Label:           DealLabel(rounded),
		DiscountAmount:  discountAmount,
		DiscountPercent: discountPercent,
	}
}

func DealLabel(score int) string {
	switch {
	case score >= 80:
		return "Deal cực tốt"
	case score >= 60:
		return "Deal tốt"
	case score >= 40:
		return "Giá ổn"
	default:
		return "Giá bình thường"
	}
}

func DiscountFromPrices(price int64, original *int64) (*int64, *float64) {
	if original == nil || *original <= price || price <= 0 {
		return nil, nil
	}
	amount := *original - price
	percent := float64(amount) / float64(*original) * 100
	return &amount, &percent
}
