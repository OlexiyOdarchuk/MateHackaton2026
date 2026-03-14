package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseAdsResponse(t *testing.T) {
	fixedNow := time.Date(2025, 1, 5, 0, 0, 0, 0, time.UTC)
	prevNow := nowFunc
	nowFunc = func() time.Time { return fixedNow }
	defer func() { nowFunc = prevNow }()

	response := `{
		"data": {
			"view_all_page": {
				"ads": {
					"edges": [
						{"node": {"id": "oldest", "start_time": "2025-01-01T00:00:00Z", "body": "Old creative", "hero_image_url": "https://cdn.example.com/old.png"}},
						{"node": {"id": "recent", "start_time": "2025-01-04T00:00:00Z", "body": "New creative", "hero_image_url": "https://cdn.example.com/new.png"}}
					]
				}
			}
		}
	}`

	creatives, err := parseAdsResponse([]byte(response))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(creatives) != 2 {
		t.Fatalf("unexpected creatives count %d", len(creatives))
	}

	if creatives[0].LongevitySeconds != 345600 {
		t.Fatalf("expected oldest longevity 345600, got %d", creatives[0].LongevitySeconds)
	}

	if creatives[1].LongevitySeconds != 86400 {
		t.Fatalf("expected recent longevity 86400, got %d", creatives[1].LongevitySeconds)
	}
}

func TestScraperScrapeTopAds(t *testing.T) {
	fixedNow := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	prevNow := nowFunc
	nowFunc = func() time.Time { return fixedNow }
	defer func() { nowFunc = prevNow }()

	lsdServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<script>{"LSD":{"token":"test-token"}}</script>`)
	}))
	defer lsdServer.Close()

	graphQLServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if r.FormValue("lsd") != "test-token" {
			t.Fatalf("unexpected LSD %q", r.FormValue("lsd"))
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		payload := `{
			"data": {
				"view_all_page": {
					"ads": {
						"edges": [
							{"node": {"id": "long", "start_time": "2025-01-02T00:00:00Z", "body": "Long running", "hero_image_url": "https://cdn.example.com/long.png"}},
							{"node": {"id": "short", "start_time": "2025-01-08T00:00:00Z", "body": "Short running", "hero_image_url": "https://cdn.example.com/short.png"}}
						]
					}
				}
			}
		}`
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, payload)
	}))
	defer graphQLServer.Close()

	s := NewScraper(nil,
		WithAdsLibraryURL(lsdServer.URL),
		WithAdsGraphQLEndpoint(graphQLServer.URL),
		WithUserAgent("unit-test-agent"),
	)

	creatives, err := s.ScrapeTopAds("page-test")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if len(creatives) != 2 {
		t.Fatalf("expected 2 creatives, got %d", len(creatives))
	}

	if creatives[0].ID != "long" {
		t.Fatalf("expected first creative 'long', got %q", creatives[0].ID)
	}

	if creatives[0].ImageURL != "https://cdn.example.com/long.png" {
		t.Fatalf("unexpected image URL %s", creatives[0].ImageURL)
	}
}
