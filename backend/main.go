package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"cloudclipboard/handlers"
	"cloudclipboard/redis"
)

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := 0
	port := getEnv("PORT", "8080")
	mode := getEnv("GIN_MODE", "release")

	gin.SetMode(mode)

	redisClient, err := redis.NewClient(redisAddr, redisPassword, redisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Println("Connected to Redis successfully")

	clipboardHandler := handlers.NewClipboardHandler(redisClient)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	frontendDir, _ := filepath.Abs("../frontend")
	indexFile := filepath.Join(frontendDir, "index.html")

	r.GET("/", func(c *gin.Context) {
		c.File(indexFile)
	})

	api := r.Group("/api")
	{
		api.GET("/health", clipboardHandler.HealthCheck)
		api.POST("/save", clipboardHandler.SaveContent)
		api.GET("/get/:code", clipboardHandler.GetContent)
	}

	r.NoRoute(func(c *gin.Context) {
		c.File(indexFile)
	})

	log.Printf("Server starting on port %s...", port)
	log.Printf("Frontend served from: %s", frontendDir)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
