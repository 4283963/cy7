package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

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
	Content string `json:"content"`
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
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "无法读取请求体: " + err.Error(),
		})
		return
	}
	defer c.Request.Body.Close()

	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求体为空，请输入要保存的内容",
		})
		return
	}

	var req SaveRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "请求格式错误，无法解析 JSON: " + err.Error(),
		})
		return
	}

	content := req.Content
	if !utf8.ValidString(content) {
		content = strings.ToValidUTF8(content, "")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "内容不能为空",
		})
		return
	}

	if len(content) > 10000 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "内容太长，最多支持 10000 字符",
		})
		return
	}

	code, err := h.redisClient.SaveContent(content, h.expiration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "保存失败: " + err.Error(),
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
			Error: "提取码格式错误，必须是 4 位数字",
		})
		return
	}

	for _, ch := range code {
		if ch < '0' || ch > '9' {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "提取码格式错误，只能包含数字",
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
