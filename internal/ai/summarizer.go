package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"SuperAdds/internal/scraper"
)

var (
	urlPattern        = regexp.MustCompile(`https?://\S+`)
	dynamicTagPattern = regexp.MustCompile(`\{\{[^}]+\}\}`)
)

const fallbackPrompt = "Cinematic, high-energy tech concept, modern aesthetics, vibrant colors, highly detailed, photorealistic, clean typography"

// SummarizeAds creates a combined English prompt from visual and textual cues.
func SummarizeAds(creatives []scraper.Creative) (string, error) {
	slog.Info("Generating aggregated insights from creatives", "count", len(creatives))

	var snippets []string
	for _, creative := range creatives {
		text := cleanDescription(creative.Description)
		if text == "" {
			continue
		}
		if dynamicTagPattern.MatchString(text) && !strings.Contains(text, "the product") {
			continue
		}
		snippets = append(snippets, text)
	}

	textSummary := buildTextSummary(snippets)

	imageURL := ""
	for _, creative := range creatives {
		if creative.ImageURL != "" {
			imageURL = creative.ImageURL
			break
		}
	}

	visionSummary := ""
	if imageURL != "" {
		if v, err := describeVisual(imageURL); err == nil {
			visionSummary = v
		} else {
			slog.Warn("Vision prompt failed, falling back", "error", err.Error())
		}
	}

	if visionSummary != "" && textSummary != "" {
		return fmt.Sprintf("%s %s", visionSummary, textSummary), nil
	}
	if visionSummary != "" {
		return visionSummary, nil
	}
	return textSummary, nil
}

func buildTextSummary(snippets []string) string {
	if len(snippets) == 0 {
		return fallbackPrompt
	}
	if len(snippets) == 1 {
		return fmt.Sprintf("%s Centered on %s.", fallbackPrompt, snippets[0])
	}
	primary := snippets[0]
	secondary := snippets[1]
	summary := fmt.Sprintf("%s Centered on %s, with supporting cues of %s.", fallbackPrompt, primary, secondary)
	if len(snippets) > 2 {
		extras := strings.Join(snippets[2:], ". ")
		summary += " " + extras
	}
	return summary
}

func cleanDescription(raw string) string {
	clean := urlPattern.ReplaceAllString(raw, "")
	clean = dynamicTagPattern.ReplaceAllString(clean, "the product")
	clean = strings.ReplaceAll(clean, "&nbsp;", " ")
	clean = strings.TrimSpace(clean)
	return clean
}

func describeVisual(imageURL string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("missing OPENAI_API_KEY")
	}

	request := map[string]interface{}{
		"model": "gpt-4o-mini",
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "input_text",
						"text": "You are an expert art director. Look at this highly successful competitor Facebook ad image. Extract the core visual concepts: composition, lighting, color palette, and subject matter. Write a highly descriptive, 2-sentence English prompt for a text-to-image AI to generate a competing ad. Focus strictly on visual elements.",
					},
					{
						"type":      "input_image",
						"image_url": imageURL,
					},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("vision API status %d: %s", resp.StatusCode, string(body))
	}

	var openResp struct {
		Choices []struct {
			Message struct {
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&openResp); err != nil {
		return "", err
	}

	for _, choice := range openResp.Choices {
		for _, c := range choice.Message.Content {
			if trimmed := strings.TrimSpace(c.Text); trimmed != "" {
				return trimmed, nil
			}
		}
	}
	return "", fmt.Errorf("vision API returned empty content")
}
