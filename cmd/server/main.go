package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"SuperAdds/internal/ai"
	"SuperAdds/internal/scraper"
	"SuperAdds/internal/store"
)

// GenerateRequest - те, що приходить від фронтенду
type GenerateRequest struct {
	UserContext string       `json:"user_context" binding:"required"`
	PageID      string       `json:"page_id" binding:"required"`
	BrandInfo   ai.BrandInfo `json:"brand_info" binding:"required"`
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

	// Налаштування Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	memStore := store.NewMemoryStore()

	// CORS мідлвар
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Хелсчек
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Ендпоінт генерації
	r.POST("/api/generate", func(c *gin.Context) {
		var req GenerateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			slog.Error("Невалідний запит від клієнта", "помилка", err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неправильний формат запиту: " + err.Error()})
			return
		}

		slog.Info("Отримано запит на генерацію", "page_id", req.PageID, "user_context", req.UserContext, "brand_description_len", len(req.BrandInfo.Description), "brand_colors", req.BrandInfo.Colors)

		// 1. Скрапимо (мокаємо) найкращі реклами
		creatives, err := scraper.ScrapeTopAds(req.PageID)
		if err != nil {
			slog.Error("Помилка скрапінгу", "помилка", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не вдалося отримати реклами"})
			return
		}

		// 2. Робимо вижимку (мокаємо)
		summary, err := ai.SummarizeAds(creatives)
		if err != nil {
			slog.Error("Помилка при створенні вижимки", "помилка", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не вдалося проаналізувати реклами"})
			return
		}

		// 3. Генеруємо нову рекламу через fal.ai
		imageURL, err := ai.GenerateAdImage(req.UserContext, summary, req.BrandInfo)
		if err != nil {
			slog.Error("Помилка генерації", "помилка", err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Помилка генерації зображення"})
			return
		}

		memStore.Save(store.StoredAd{
			PageID:      req.PageID,
			UserContext: req.UserContext,
			BrandInfo:   req.BrandInfo,
			Summary:     summary,
			ImageURL:    imageURL,
		})

		slog.Info("Успішно згенеровано рекламу", "image_url", imageURL)

		c.JSON(http.StatusOK, GenerateResponse{
			ImageURL: imageURL,
			Summary:  summary,
		})
	})

	r.GET("/api/store/:page_id", func(c *gin.Context) {
		pageID := c.Param("page_id")
		if stored, ok := memStore.Get(pageID); ok {
			c.JSON(http.StatusOK, stored)
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Дані для цієї сторінки не знайдено"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Сервер запущено", "port", port)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		slog.Error("Помилка запуску сервера", "помилка", err.Error())
	}
}
