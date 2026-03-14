package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
)

// FalResponse - структура для парсингу відповіді від fal.ai
type FalResponse struct {
	Images []struct {
		URL string `json:"url"`
	} `json:"images"`
}

// GenerateAdImage викликає fal-ai/nano-banana-2 для генерації нової реклами 
// на основі вводу користувача та вижимки з найкращих реклам конкурентів.
func GenerateAdImage(userContext, adSummary string) (string, error) {
	falKey := os.Getenv("FAL_KEY")
	if falKey == "" {
		return "", fmt.Errorf("FAL_KEY не знайдено в змінних середовища")
	}

	// Формуємо фінальний промпт, поєднуючи побажання юзера і вижимку
	finalPrompt := fmt.Sprintf("Create a highly engaging advertisement image. User request: %s. Use these successful elements from competitors: %s. High quality, professional, photorealistic.", userContext, adSummary)
	
	slog.Info("Генерація фото через fal.ai", "prompt_length", len(finalPrompt))

	// Тіло запиту
	reqBody, _ := json.Marshal(map[string]string{
		"prompt": finalPrompt,
	})

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
