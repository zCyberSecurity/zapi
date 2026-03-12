package handler

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zCyberSecurity/zapi/internal/middleware"
	"github.com/zCyberSecurity/zapi/internal/model"
	"github.com/zCyberSecurity/zapi/internal/proxy"
	"gorm.io/gorm"
)

type AnthropicHandler struct {
	db    *gorm.DB
	proxy *proxy.Proxy
}

func NewAnthropicHandler(db *gorm.DB) *AnthropicHandler {
	return &AnthropicHandler{db: db, proxy: proxy.New(db)}
}

func (h *AnthropicHandler) CreateMessage(c *gin.Context) {
	apiKey := c.MustGet(middleware.CtxAPIKey).(*model.APIKey)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("failed to read request body"))
		return
	}

	// Parse the Anthropic request to get model and stream flag.
	var areq proxy.AnthropicRequest
	if err := unmarshalJSON(body, &areq); err != nil {
		c.JSON(http.StatusBadRequest, errResp("invalid request body"))
		return
	}
	if areq.Model == "" {
		c.JSON(http.StatusBadRequest, errResp("model is required"))
		return
	}
	if !apiKey.HasModelAccess(areq.Model) {
		c.JSON(http.StatusForbidden, errResp("API key does not have access to model: "+areq.Model))
		return
	}

	provider, pm, err := h.proxy.FindProvider(areq.Model)
	if err != nil {
		c.JSON(http.StatusNotFound, errResp(err.Error()))
		return
	}

	// If provider is native Anthropic, forward as-is.
	if provider.APIType == "anthropic" {
		upstreamBody := proxy.ReplaceModel(body, pm.UpstreamModelID())
		if areq.Stream {
			h.proxy.ForwardStream(c.Writer, c.Request, provider, "/messages", upstreamBody) //nolint
			return
		}
		status, respBody, err := h.proxy.ForwardBuffered(c.Request, provider, "/messages", upstreamBody)
		if err != nil {
			c.JSON(http.StatusBadGateway, errResp(err.Error()))
			return
		}
		if status == http.StatusOK {
			var ar proxy.AnthropicResponse
			if unmarshalJSON(respBody, &ar) == nil && ar.Usage.OutputTokens > 0 {
				h.recordUsage(apiKey, areq.Model,
					ar.Usage.InputTokens, ar.Usage.OutputTokens,
					ar.Usage.InputTokens+ar.Usage.OutputTokens)
			}
		}
		c.Data(status, "application/json", respBody)
		return
	}

	// Convert Anthropic → OpenAI, forward, convert response back.
	oaiBody, err := proxy.AnthropicToOpenAI(body)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	oaiBody = proxy.ReplaceModel(oaiBody, pm.UpstreamModelID())

	if areq.Stream {
		streamBody := proxy.InjectStreamOptions(oaiBody)
		prompt, completion, total, _ := h.proxy.ForwardStream(c.Writer, c.Request, provider, "/chat/completions", streamBody)
		if total > 0 {
			h.recordUsage(apiKey, areq.Model, prompt, completion, total)
		}
		return
	}

	status, respBody, err := h.proxy.ForwardBuffered(c.Request, provider, "/chat/completions", oaiBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, errResp(err.Error()))
		return
	}
	if status != http.StatusOK {
		c.Data(status, "application/json", respBody)
		return
	}

	prompt, completion, total := proxy.ExtractOpenAIUsage(respBody)
	if total > 0 {
		h.recordUsage(apiKey, areq.Model, prompt, completion, total)
	}

	converted, err := proxy.OpenAIToAnthropic(respBody)
	if err != nil {
		c.Data(status, "application/json", respBody)
		return
	}
	c.Data(http.StatusOK, "application/json", converted)
}

func (h *AnthropicHandler) recordUsage(apiKey *model.APIKey, modelID string, prompt, completion, total int) {
	date := time.Now().Format("2006-01-02")
	var usage model.UsageLog
	err := h.db.Where("api_key_id = ? AND model = ? AND date = ?", apiKey.ID, modelID, date).First(&usage).Error
	if err != nil {
		usage = model.UsageLog{
			APIKeyID:   apiKey.ID,
			APIKeyName: apiKey.Name,
			Model:      modelID,
			Date:       date,
		}
		h.db.Create(&usage)
	}
	h.db.Model(&usage).Updates(map[string]interface{}{
		"prompt_tokens":     usage.PromptTokens + prompt,
		"completion_tokens": usage.CompletionTokens + completion,
		"total_tokens":      usage.TotalTokens + total,
		"request_count":     usage.RequestCount + 1,
	})
}
