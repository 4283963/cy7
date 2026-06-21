package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"cloudclipboard/redis"
)

type ClipboardHandler struct {
	redisClient *redis.Client
	expiration  time.Duration
}

func NewClipboardHandler(redisClient *redis.Client) *ClipboardHandler {
	return &ClipboardHandler{
		redisClient: redisClient,
		expiration:  24 * time.Hour,
	}
}

type SaveRequest struct {
	Content string `json:"content" binding:"required,min=1,max=10000"`
}

type SaveResponse struct {
	Code string `json:"code"`
}

type GetResponse struct {
	Content string `json:"content"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *ClipboardHandler) SaveContent(c *gin.Context) {
	var req SaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body. Content must be 1-10000 characters.",
		})
		return
	}

	code, err := h.redisClient.SaveContent(req.Content, h.expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to save content: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SaveResponse{
		Code: code,
	})
}

func (h *ClipboardHandler) GetContent(c *gin.Context) {
	code := c.Param("code")

	if len(code) != 4 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid code format. Code must be 4 digits.",
		})
		return
	}

	for _, ch := range code {
		if ch < '0' || ch > '9' {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid code format. Code must contain only digits.",
			})
			return
		}
	}

	content, err := h.redisClient.GetContent(code)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetResponse{
		Content: content,
	})
}

func (h *ClipboardHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
