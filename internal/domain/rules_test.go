package domain

import "testing"

func TestDealScoreHighWhenPriceIsHistoricallyLow(t *testing.T) {
	original := int64(200)
	result := CalculateDealScore(DealScoreInput{
		CurrentPrice:    100,
		OriginalPrice:   &original,
		Avg30:           1000,
		Min30:           100,
		HistoricalMin:   100,
		MerchantQuality: 5,
	})
	if result.Score < 80 {
		t.Fatalf("score = %d, want at least 80", result.Score)
	}
	if result.Label != "Deal cực tốt" {
		t.Fatalf("label = %s", result.Label)
	}
	if result.DiscountAmount == nil || *result.DiscountAmount != 100 {
		t.Fatalf("discount amount = %v", result.DiscountAmount)
	}
}

func TestDealScoreIgnoresDiscountWhenOriginalIsNotHigher(t *testing.T) {
	original := int64(1000)
	result := CalculateDealScore(DealScoreInput{
		CurrentPrice:    1000,
		OriginalPrice:   &original,
		MerchantQuality: 0,
	})
	if result.DiscountAmount != nil || result.DiscountPercent != nil {
		t.Fatalf("expected no discount, got amount %v percent %v", result.DiscountAmount, result.DiscountPercent)
	}
	if result.Score != 0 {
		t.Fatalf("score = %d, want 0", result.Score)
	}
	if result.Label != "Giá bình thường" {
		t.Fatalf("label = %s", result.Label)
	}
}

func TestDealLabels(t *testing.T) {
	cases := map[int]string{0: "Giá bình thường", 39: "Giá bình thường", 40: "Giá ổn", 59: "Giá ổn", 60: "Deal tốt", 79: "Deal tốt", 80: "Deal cực tốt", 100: "Deal cực tốt"}
	for score, label := range cases {
		if got := DealLabel(score); got != label {
			t.Fatalf("score %d label %s, want %s", score, got, label)
		}
	}
}

func TestSummarizeOffers(t *testing.T) {
	offers := []Offer{
		{Price: 300, Status: StatusActive},
		{Price: 100, Status: StatusActive},
		{Price: 200, Status: StatusInactive},
	}
	lowest, average, savings, _ := SummarizeOffers(offers)
	if lowest == nil || *lowest != 100 {
		t.Fatalf("lowest = %v", lowest)
	}
	if average == nil || *average != 200 {
		t.Fatalf("average = %v", average)
	}
	if savings == nil || *savings != 100 {
		t.Fatalf("savings = %v", savings)
	}
}

func TestHistoricalPriceTodayMatchesCurrent(t *testing.T) {
	if got := HistoricalPrice(22190000, 0, true); got != 22190000 {
		t.Fatalf("today price = %d", got)
	}
	if got := HistoricalPrice(22190000, 14, false); got <= 22190000 {
		t.Fatalf("past price without dip should be higher, got %d", got)
	}
}

func TestNormalizePageCapsLimit(t *testing.T) {
	page, limit := NormalizePage(0, 500)
	if page != 1 || limit != 100 {
		t.Fatalf("page %d limit %d", page, limit)
	}
}

func TestSlugifyVietnamese(t *testing.T) {
	if got := Slugify("Điện Máy Xanh"); got != "dien-may-xanh" {
		t.Fatalf("slug = %s", got)
	}
}
