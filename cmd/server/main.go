package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"SuperAdds/internal/ai"
	"SuperAdds/internal/scraper"
	"SuperAdds/internal/store"
)

// GenerateRequest - те, що приходить від фронтенду
type GenerateRequest struct {
	UserContext    string       `json:"user_context" binding:"required"`
	CompetitorIDs  []string     `json:"competitor_ids" binding:"required"`
	AdLanguage     string       `json:"ad_language"`
	BrandInfo      ai.BrandInfo `json:"brand_info" binding:"required"`
}

// GenerateResponse - те, що ми повертаємо на фронтенд
type GenerateResponse struct {
	ImageURL string `json:"image_url"`
	Summary  string `json:"summary"`
}

func main() {
	// Налаштування slog як головного логера
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Завантаження .env файлу
	if err := godotenv.Load(); err != nil {
		slog.Warn("Файл .env не знайдено або помилка читання. Використовуємо системні змінні.")
	} else {
		slog.Info("Завантажено змінні середовища з .env файлу.")
	}

	r := setupRouter(store.NewMemoryStore())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Сервер запущено", "port", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		slog.Error("Помилка запуску сервера", "помилка", err.Error())
	}
}

func setupRouter(memStore *store.MemoryStore) *gin.Engine {
	r := gin.Default()

	// CORS мідлвар
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/api/generate", generateAdHandler(memStore))
	r.GET("/api/store/:page_id", storeLookupHandler(memStore))

	return r
}

func generateAdHandler(memStore *store.MemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("Невалідний запит від клієнта", "помилка", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильний формат запиту: " + err.Error()})
			return
		}

		pageIDs := derivePageIDs(req.CompetitorIDs)
		if len(pageIDs) == 0 {
			msg := "немає валідних competitor_ids"
			slog.Warn("Немає competitor IDs", "input", req.CompetitorIDs)
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		adLanguage := strings.TrimSpace(req.AdLanguage)
		if adLanguage == "" {
			adLanguage = "Ukrainian"
		}

		slog.Info(
			"Отримано запит на генерацію",
			"competitors", pageIDs,
			"user_context", req.UserContext,
			"ad_language", adLanguage,
			"brand_description_len", len(req.BrandInfo.Description),
			"brand_colors", req.BrandInfo.Colors,
		)

		slog.Info("Merging data from competitors", "count", len(pageIDs))

		var (
			wg           sync.WaitGroup
			mu           sync.Mutex
			allCreatives []scraper.Creative
			errFirst     error
		)
		errCh := make(chan error, len(pageIDs))

		for _, pageID := range pageIDs {
			wg.Add(1)
			go func(pid string) {
				defer wg.Done()
				scraperInstance := scraper.NewScraper(nil)
				creatives, err := scraperInstance.ScrapeTopAds(pid)
				if err != nil {
					errCh <- fmt.Errorf("page %s: %w", pid, err)
					return
				}
				if len(creatives) == 0 {
					slog.Warn("Скрап повернув 0 креативів", "page_id", pid)
				}
				mu.Lock()
				allCreatives = append(allCreatives, creatives...)
				mu.Unlock()
			}(pageID)
		}

		wg.Wait()
		close(errCh)
		for err := range errCh {
			if errFirst == nil {
				errFirst = err
			}
			slog.Error("Помилка під час скрапінгу конкурента", "error", err.Error())
		}

		if len(allCreatives) == 0 {
			slog.Warn("Scraper returned zero creatives, continuing with demo summary", "error", errFirst)
		}

		if output, err := json.MarshalIndent(allCreatives, "", "  "); err == nil {
			if err := os.WriteFile("debug_scraper_output.json", output, 0o644); err != nil {
				slog.Warn("Unable to write debug scraper dump", "error", err.Error())
			} else {
				slog.Info("Wrote debug scraper dump", "file", "debug_scraper_output.json")
			}
		} else {
			slog.Warn("Unable to marshal creatives for debug dump", "error", err.Error())
		}

		summary, err := ai.SummarizeAds(allCreatives)
		if err != nil {
			slog.Error("Помилка при створенні вижимки", "помилка", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не вдалося проаналізувати реклами"})
			return
		}

		imageURL, err := ai.GenerateAdImage(req.UserContext, summary, adLanguage, req.BrandInfo)
		if err != nil {
			slog.Error("Помилка генерації", "помилка", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Помилка генерації зображення"})
			return
		}

		memStore.Save(store.StoredAd{
			PageID:      strings.Join(pageIDs, ","),
			UserContext: req.UserContext,
			BrandInfo:   req.BrandInfo,
			Competitors: pageIDs,
			Summary:     summary,
			ImageURL:    imageURL,
		})

		slog.Info("Успішно згенеровано рекламу", "image_url", imageURL)

		c.JSON(http.StatusOK, GenerateResponse{
			ImageURL: imageURL,
			Summary:  summary,
		})
	}
}

func storeLookupHandler(memStore *store.MemoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageID := c.Param("page_id")
		if stored, ok := memStore.Get(pageID); ok {
			c.JSON(http.StatusOK, stored)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Дані для цієї сторінки не знайдено"})
	}
}

func derivePageIDs(inputs []string) []string {
	result := make([]string, 0, len(inputs))
	seen := map[string]struct{}{}
	for _, input := range inputs {
		if pid := extractPageID(input); pid != "" {
			if _, ok := seen[pid]; ok {
				continue
			}
			seen[pid] = struct{}{}
			result = append(result, pid)
		}
	}
	return result
}

func extractPageID(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http") {
		parsed, err := url.Parse(raw)
		if err == nil {
			if pid := parsed.Query().Get("view_all_page_id"); pid != "" {
				return pid
			}
			if pid := parsed.Query().Get("page_id"); pid != "" {
				return pid
			}
			path := strings.Trim(parsed.Path, "/")
			if path != "" {
				segments := strings.Split(path, "/")
				return segments[len(segments)-1]
			}
		}
	}
	return raw
}
