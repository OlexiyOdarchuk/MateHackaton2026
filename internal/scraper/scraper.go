package scraper

// Creative - структура, що зберігає дані про креатив у пам'яті
type Creative struct {
	ImageURL    string
	ImageBytes  []byte
	Description string
	Duration    int64 // Тривалість роботи реклами (в секундах)
}

// ScrapeTopAds мокована версія скрапера.
func ScrapeTopAds(pageID string) ([]Creative, error) {
	// Повертаємо 5 мокових креативів
	creatives := []Creative{}
	for i := 1; i <= 5; i++ {
		creatives = append(creatives, Creative{
			ImageURL:    "https://example.com/mock-image.jpg",
			ImageBytes:  []byte("mock image data"),
			Description: "Це найкраща реклама номер " + string(rune('0'+i)) + "! Купуйте наші товари.",
			Duration:    int64(i * 86400 * 10), // 10, 20, 30... днів
		})
	}
	return creatives, nil
}
