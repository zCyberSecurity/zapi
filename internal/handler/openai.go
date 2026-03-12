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

type OpenAIHandler struct {
	db    *gorm.DB
	proxy *proxy.Proxy
}

func NewOpenAIHandler(db *gorm.DB) *OpenAIHandler {
	return &OpenAIHandler{db: db, proxy: proxy.New(db)}
}

func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	apiKey := c.MustGet(middleware.CtxAPIKey).(*model.APIKey)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("failed to read request body"))
		return
	}

	req := proxy.ParseOpenAIRequest(body)
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, errResp("model is required"))
		return
	}
	if !apiKey.HasModelAccess(req.Model) {
		c.JSON(http.StatusForbidden, errResp("API key does not have access to model: "+req.Model))
		return
	}

	provider, pm, err := h.proxy.FindProvider(req.Model)
	if err != nil {
		c.JSON(http.StatusNotFound, errResp(err.Error()))
		return
	}

	// Replace model ID with the upstream model ID if different.
	upstreamBody := proxy.ReplaceModel(body, pm.UpstreamModelID())

	if req.Stream {
		h.proxy.ForwardStream(c.Writer, c.Request, provider, "/chat/completions", upstreamBody)
		return
	}

	status, respBody, err := h.proxy.ForwardBuffered(c.Request, provider, "/chat/completions", upstreamBody)
	if err != nil {
		c.JSON(http.StatusBadGateway, errResp(err.Error()))
		return
	}

	if status == http.StatusOK {
		prompt, completion, total := proxy.ExtractOpenAIUsage(respBody)
		if total > 0 {
			h.recordUsage(apiKey, req.Model, prompt, completion, total)
		}
	}

	c.Data(status, "application/json", respBody)
}

func (h *OpenAIHandler) ListModels(c *gin.Context) {
	apiKey := c.MustGet(middleware.CtxAPIKey).(*model.APIKey)

	var pms []model.ProviderModel
	h.db.Preload("Provider").
		Joins("JOIN providers ON providers.id = provider_models.provider_id AND providers.enabled = true AND providers.deleted_at IS NULL").
		Where("provider_models.enabled = ? AND provider_models.deleted_at IS NULL", true).
		Find(&pms)

	type modelObj struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		OwnedBy string `json:"owned_by"`
	}

	var data []modelObj
	for _, pm := range pms {
		if apiKey.HasModelAccess(pm.ModelID) {
			data = append(data, modelObj{
				ID:      pm.ModelID,
				Object:  "model",
				Created: time.Now().Unix(),
				OwnedBy: pm.Provider.Name,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"object": "list", "data": data})
}

func (h *OpenAIHandler) recordUsage(apiKey *model.APIKey, modelID string, prompt, completion, total int) {
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
