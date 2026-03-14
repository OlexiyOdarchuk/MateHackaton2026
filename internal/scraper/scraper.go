package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	defaultAdsLibraryURL        = "https://www.facebook.com/ads/library/"
	defaultAdsGraphQLEndpoint   = "https://www.facebook.com/api/graphql/"
	defaultUserAgent            = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"
	graphQLDocID                = "26738502115734848"
	maxReturnedCreatives        = 5
	defaultClientTimeoutSeconds = 30
)

var lsdTokenPattern = regexp.MustCompile(`"LSD"\s*:\s*\{\s*"token"\s*:\s*"([^"]+)"`)

var nowFunc = func() time.Time {
	return time.Now().UTC()
}

var jitterSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// ScraperOption mutates scraper settings.
type ScraperOption func(*Scraper)

// WithAdsLibraryURL overrides the default Facebook ads library URL.
func WithAdsLibraryURL(u string) ScraperOption {
	return func(s *Scraper) {
		s.adsLibraryURL = u
	}
}

// WithAdsGraphQLEndpoint overrides the default GraphQL endpoint.
func WithAdsGraphQLEndpoint(u string) ScraperOption {
	return func(s *Scraper) {
		s.adsGraphQLEndpoint = u
	}
}

// WithUserAgent overrides the user-agent string.
func WithUserAgent(agent string) ScraperOption {
	return func(s *Scraper) {
		s.userAgent = agent
	}
}

// Scraper owns the client and endpoint settings.
type Scraper struct {
	client              *http.Client
	adsLibraryURL       string
	adsGraphQLEndpoint  string
	userAgent           string
}

// Creative carries the essential parsed ad information.
type Creative struct {
	ID               string    `json:"id"`
	ImageURL         string    `json:"image_url"`
	Description      string    `json:"description"`
	StartTime        time.Time `json:"start_time"`
	LongevitySeconds int64     `json:"longevity_seconds"`
}

// ScrapeTopAds returns the top 5 longest-running Facebook ads for the supplied page.
func ScrapeTopAds(pageID string) ([]Creative, error) {
	return NewScraper(nil).ScrapeTopAds(pageID)
}

// NewScraper creates a scraper service with optional overrides.
func NewScraper(client *http.Client, opts ...ScraperOption) *Scraper {
	if client == nil {
		jar, _ := cookiejar.New(nil)
		client = &http.Client{
			Jar:     jar,
			Timeout: defaultClientTimeoutSeconds * time.Second,
		}
	}

	s := &Scraper{
		client:             client,
		adsLibraryURL:      defaultAdsLibraryURL,
		adsGraphQLEndpoint: defaultAdsGraphQLEndpoint,
		userAgent:          defaultUserAgent,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ScrapeTopAds returns the top creatives.
func (s *Scraper) ScrapeTopAds(pageID string) ([]Creative, error) {
	if pageID == "" {
		return nil, errors.New("page_id is required")
	}

	token, manualCookie, err := s.fetchLSDToken()
	if err != nil {
		return nil, err
	}

	creatives, err := s.fetchAds(token, pageID, manualCookie)
	if err != nil {
		return nil, fmt.Errorf("fetch ads: %w", err)
	}

	return creatives, nil
}

func (s *Scraper) fetchLSDToken() (string, string, error) {
	req, err := http.NewRequest(http.MethodGet, s.adsLibraryURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to build LSD request: %w", err)
	}
	s.applyGetHeaders(req)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("LSD request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		slog.Warn("Auto-fetch failed with 403, falling back to manual env vars")
		return fallbackTokenFromEnv()
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("LSD request returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read LSD response: %w", err)
	}

	match := lsdTokenPattern.FindSubmatch(body)
	if len(match) < 2 {
		slog.Warn("Auto-fetch failed to parse LSD token, falling back to manual env vars")
		return fallbackTokenFromEnv()
	}

	token := string(match[1])
	slog.Debug("LSD token extracted", "length", len(token))
	return token, "", nil
}

func (s *Scraper) fetchAds(token, pageID, manualCookie string) ([]Creative, error) {
	vars := map[string]interface{}{
		"activeStatus": "ACTIVE",
		"adType":       "ALL",
		"countries":    []string{"UA"},
		"viewAllPageID": pageID,
		"sortData": map[string]interface{}{
			"mode":      "SORT_BY_TOTAL_IMPRESSIONS",
			"direction": "DESCENDING",
		},
	}

	varsBytes, err := json.Marshal(vars)
	if err != nil {
		return nil, fmt.Errorf("failed to build variables payload: %w", err)
	}

	values := url.Values{}
	values.Set("doc_id", graphQLDocID)
	values.Set("lsd", token)
	values.Set("variables", string(varsBytes))

	req, err := http.NewRequest(http.MethodPost, s.adsGraphQLEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to build GraphQL request: %w", err)
	}
	s.applyPostHeaders(req, token, manualCookie, pageID)

	jitterSleep()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GraphQL request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("403 forbidden from Facebook; consider rotating cookies or using a proxy")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected GraphQL status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read GraphQL response: %w", err)
	}

	creatives, err := parseAdsResponse(body)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(creatives, func(i, j int) bool {
		return creatives[i].LongevitySeconds > creatives[j].LongevitySeconds
	})

	if len(creatives) > maxReturnedCreatives {
		creatives = creatives[:maxReturnedCreatives]
	}

	return creatives, nil
}

func (s *Scraper) applyGetHeaders(req *http.Request) {
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
}

func (s *Scraper) applyPostHeaders(req *http.Request, token, manualCookie, pageID string) {
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-FB-LSD", token)
	req.Header.Set("Origin", "https://www.facebook.com")
	req.Header.Set("Referer", fmt.Sprintf("%s?active_status=all&ad_type=all&view_all_page_id=%s", s.adsLibraryURL, pageID))
	if manualCookie != "" {
		req.Header.Set("Cookie", manualCookie)
	}
}

func jitterSleep() {
	ms := 500 + jitterSource.Intn(2000)
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func fallbackTokenFromEnv() (string, string, error) {
	token := strings.TrimSpace(os.Getenv("FB_LSD_TOKEN"))
	cookie := strings.TrimSpace(os.Getenv("FB_COOKIE"))
	if token == "" {
		return "", "", errors.New("FB_LSD_TOKEN env var is required while falling back")
	}
	return token, cookie, nil
}

func parseAdsResponse(body []byte) ([]Creative, error) {
	if edges := extractEdgesFromJSON(body); len(edges) > 0 {
		return creativesFromEdges(edges)
	}

	var edges []interface{}
	lines := bytes.Split(body, []byte("\n"))
	for _, lineBytes := range lines {
		line := bytes.TrimSpace(lineBytes)
		if len(line) == 0 {
			continue
		}
		if edges = extractEdgesFromJSON(line); len(edges) > 0 {
			break
		}
	}
	if len(edges) == 0 {
		return nil, errors.New("no creatives parsed; the GraphQL structure may have changed")
	}

	return creativesFromEdges(edges)
}

func creativesFromEdges(edges []interface{}) ([]Creative, error) {
	var creatives []Creative
	now := nowFunc()
	for _, rawEdge := range edges {
		edgeMap, ok := rawEdge.(map[string]interface{})
		if !ok {
			continue
		}

		node, ok := edgeMap["node"].(map[string]interface{})
		if !ok {
			continue
		}

		startStr, _ := findStringByKeys(node, []string{
			"start_time", "start_date", "ad_start_time", "delivery_start_time",
		})

		startTime := parseStartTime(startStr)

		imageURL, _ := findImageURL(node)
		description := extractDescription(node)

		creative := Creative{
			ID:          findStringByKeysOrEmpty(node, "id"),
			ImageURL:    imageURL,
			Description: description,
			StartTime:   startTime,
		}
		if !startTime.IsZero() {
			creative.LongevitySeconds = int64(now.Sub(startTime).Seconds())
			if creative.LongevitySeconds < 0 {
				creative.LongevitySeconds = 0
			}
		}

		creatives = append(creatives, creative)
	}

	if len(creatives) == 0 {
		return nil, errors.New("no creatives parsed; the GraphQL structure may have changed")
	}

	return creatives, nil
}

func extractEdgesFromJSON(raw []byte) []interface{} {
	var chunk map[string]interface{}
	if err := json.Unmarshal(raw, &chunk); err != nil {
		return nil
	}
	data, ok := chunk["data"].(map[string]interface{})
	if !ok {
		return nil
	}
	return findEdges(data)
}

func findEdges(data map[string]interface{}) []interface{} {
	if viewPage, ok := data["view_all_page"].(map[string]interface{}); ok {
		if adsSection, ok := viewPage["ads"].(map[string]interface{}); ok {
			if edges, ok := adsSection["edges"].([]interface{}); ok {
				return edges
			}
		}
	}
	if adLibrary, ok := data["ad_library_main"].(map[string]interface{}); ok {
		if searchConn, ok := adLibrary["search_results_connection"].(map[string]interface{}); ok {
			if edges, ok := searchConn["edges"].([]interface{}); ok {
				return edges
			}
		}
	}
	return nil
}

func extractDescription(node map[string]interface{}) string {
	var fragments []string
	if snapshot, ok := node["snapshot"].(map[string]interface{}); ok {
		appendIfNotEmpty(&fragments, extractSnapshotText(snapshot))
		if cards, ok := snapshot["cards"].([]interface{}); ok {
			for _, rawCard := range cards {
				if card, ok := rawCard.(map[string]interface{}); ok {
					appendIfNotEmpty(&fragments, extractCardText(card))
				}
			}
		}
	}
	if len(fragments) == 0 {
		if fallback, ok := findStringByKeys(node, []string{"body", "ad_text", "headline", "title", "message"}); ok {
			appendIfNotEmpty(&fragments, fallback)
		}
	}
	return strings.Join(fragments, " ")
}

func extractSnapshotText(snapshot map[string]interface{}) string {
	var pieces []string
	if bodyVal, ok := snapshot["body"]; ok {
		if bodyMap, ok := bodyVal.(map[string]interface{}); ok {
			if txt, ok := nestedString(bodyMap, "text"); ok {
				appendIfNotEmpty(&pieces, txt)
			}
		} else if txt, ok := bodyVal.(string); ok {
			appendIfNotEmpty(&pieces, txt)
		}
	}
	appendIfNotEmpty(&pieces, valueAsString(snapshot, "title"))
	appendIfNotEmpty(&pieces, valueAsString(snapshot, "caption"))
	appendIfNotEmpty(&pieces, valueAsString(snapshot, "message"))
	return strings.Join(pieces, " ")
}

func extractCardText(card map[string]interface{}) string {
	var fragments []string
	appendIfNotEmpty(&fragments, nestedStringValue(card, []string{"body", "text"}))
	appendIfNotEmpty(&fragments, valueAsString(card, "title"))
	appendIfNotEmpty(&fragments, valueAsString(card, "caption"))
	appendIfNotEmpty(&fragments, valueAsString(card, "message"))
	return strings.Join(fragments, " ")
}

func nestedStringValue(root map[string]interface{}, path []string) string {
	if txt, ok := nestedString(root, path...); ok {
		return txt
	}
	return ""
}

func nestedString(root map[string]interface{}, path ...string) (string, bool) {
	current := root
	for i, key := range path {
		if i == len(path)-1 {
			if val, ok := current[key]; ok {
				if s, ok := val.(string); ok && strings.TrimSpace(s) != "" {
					return s, true
				}
			}
			return "", false
		}
		next, ok := current[key].(map[string]interface{})
		if !ok {
			return "", false
		}
		current = next
	}
	return "", false
}

func valueAsString(node map[string]interface{}, key string) string {
	if val, ok := node[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func appendIfNotEmpty(list *[]string, candidate string) {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return
	}
	for _, existing := range *list {
		if strings.EqualFold(existing, candidate) {
			return
		}
	}
	*list = append(*list, candidate)
}

func parseStartTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05+0000",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}

func findStringByKeys(node interface{}, keys []string) (string, bool) {
	switch typed := node.(type) {
	case map[string]interface{}:
		for k, v := range typed {
			for _, matchKey := range keys {
				if strings.EqualFold(k, matchKey) {
					if s, ok := v.(string); ok && s != "" {
						return s, true
					}
				}
			}
			if s, ok := findStringByKeys(v, keys); ok {
				return s, true
			}
		}
	case []interface{}:
		for _, item := range typed {
			if s, ok := findStringByKeys(item, keys); ok {
				return s, true
			}
		}
	}
	return "", false
}

func findStringByKeysOrEmpty(node map[string]interface{}, key string) string {
	if val, ok := node[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

func findImageURL(node interface{}) (string, bool) {
	switch typed := node.(type) {
	case map[string]interface{}:
		for k, v := range typed {
			if s, ok := v.(string); ok && looksLikeImageURL(s) {
				if strings.Contains(strings.ToLower(k), "image") || strings.Contains(strings.ToLower(k), "media") {
					return s, true
				}
			}
			if s, ok := findImageURL(v); ok {
				return s, true
			}
		}
	case []interface{}:
		for _, item := range typed {
			if s, ok := findImageURL(item); ok {
				return s, true
			}
		}
	}
	return "", false
}

func looksLikeImageURL(value string) bool {
	if !strings.HasPrefix(value, "http") {
		return false
	}
	lower := strings.ToLower(value)
	return strings.Contains(lower, ".png") || strings.Contains(lower, ".jpg") || strings.Contains(lower, ".jpeg") || strings.Contains(lower, ".webp")
}
