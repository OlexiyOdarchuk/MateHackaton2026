package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// BrandInfo describes the brand identity assets that the frontend sends.
type BrandInfo struct {
	LogoImage      string   `json:"logo_image" binding:"required"`
	Description    string   `json:"company_description" binding:"required"`
	Colors         []string `json:"company_colors"`
	CreativePrompt string   `json:"creative_prompt"`
}

// FalResponse - структура для парсингу відповіді від fal.ai
type FalResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

// GenerateAdImage викликає fal-ai/nano-banana-2 для генерації нової реклами
// на основі вводу користувача, бренду та вижимки з найкращих реклам конкурентів.
func GenerateAdImage(userContext, adSummary string, brand BrandInfo) (string, error) {
	falKey := os.Getenv("FAL_KEY")
	if falKey == "" {
		mockURL := "https://example.com/mock-ad-image.png"
		slog.Warn("FAL_KEY не задано, повертаємо мокове зображення для MVP", "mock_url", mockURL)
		return mockURL, nil
	}

	// Формуємо фінальний промпт, поєднуючи побажання юзера і вижимку
	colorPalette := "палітра ще не задана"
	if len(brand.Colors) > 0 {
		colorPalette = strings.Join(brand.Colors, ", ")
	}

	creativeHint := brand.CreativePrompt
	if creativeHint == "" {
		creativeHint = "Підкресли цінність продукту, встанови емоційний тон і запропонуй сильний заклик до дії."
	}

	logoNote := fmt.Sprintf("Лого надано як SVG/PNG-дані (довжина %d символів).", len(brand.LogoImage))

	finalPrompt := fmt.Sprintf(
		"Create a dramatic, differentiated advertisement image. User request: %s. Brand description: %s. %s Colors: %s. Creative prompt: %s. Competitor insights: %s. Include the provided logo as the primary lockup and honour the brand palette, but add your own elevated visual storytelling. High quality, professional, photorealistic.",
		userContext,
		brand.Description,
		logoNote,
		colorPalette,
		creativeHint,
		adSummary,
	)

	slog.Info("Генерація фото через fal.ai", "prompt_length", len(finalPrompt))

	// Тіло запиту
	reqBody, _ := json.Marshal(map[string]string{
		"prompt": finalPrompt,
	})
	slog.Info("Ось final prompt", "prompt", finalPrompt)
	url := "https://queue.fal.run/fal-ai/nano-banana-2"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("помилка створення запиту до fal.ai: %w", err)
	}

	req.Header.Set("Authorization", "Key "+falKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("помилка виконання запиту до fal.ai: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("помилка fal.ai: статус %d, тіло: %s", resp.StatusCode, string(body))
	}

	var falResp FalResponse
	if err := json.NewDecoder(resp.Body).Decode(&falResp); err != nil {
		return "", fmt.Errorf("помилка парсингу відповіді fal.ai: %w", err)
	}

	if len(falResp.Images) > 0 {
		return falResp.Images[0].URL, nil
	}

	return "", fmt.Errorf("fal.ai не повернув жодного зображення")
}
