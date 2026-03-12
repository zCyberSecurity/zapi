package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zCyberSecurity/zapi/internal/model"
	"gorm.io/gorm"
)

type AdminHandler struct {
	db *gorm.DB
}

func NewAdminHandler(db *gorm.DB) *AdminHandler {
	return &AdminHandler{db: db}
}

// --- Providers ---

func (h *AdminHandler) ListProviders(c *gin.Context) {
	var providers []model.Provider
	h.db.Preload("Models").Find(&providers)
	c.JSON(http.StatusOK, providers)
}

func (h *AdminHandler) CreateProvider(c *gin.Context) {
	var p model.Provider
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	if err := h.db.Create(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *AdminHandler) UpdateProvider(c *gin.Context) {
	var p model.Provider
	if err := h.db.First(&p, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errResp("provider not found"))
		return
	}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	h.db.Save(&p)
	c.JSON(http.StatusOK, p)
}

func (h *AdminHandler) DeleteProvider(c *gin.Context) {
	if err := h.db.Delete(&model.Provider{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Models ---

func (h *AdminHandler) ListModels(c *gin.Context) {
	var models []model.ProviderModel
	h.db.Where("provider_id = ?", c.Param("id")).Find(&models)
	c.JSON(http.StatusOK, models)
}

func (h *AdminHandler) AddModel(c *gin.Context) {
	providerID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errResp("invalid provider id"))
		return
	}
	var pm model.ProviderModel
	if err := c.ShouldBindJSON(&pm); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	pm.ProviderID = uint(providerID)
	if err := h.db.Create(&pm).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, pm)
}

func (h *AdminHandler) UpdateModel(c *gin.Context) {
	var pm model.ProviderModel
	if err := h.db.First(&pm, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errResp("model not found"))
		return
	}
	if err := c.ShouldBindJSON(&pm); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	h.db.Save(&pm)
	c.JSON(http.StatusOK, pm)
}

func (h *AdminHandler) DeleteModel(c *gin.Context) {
	if err := h.db.Delete(&model.ProviderModel{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// --- API Keys ---

func (h *AdminHandler) ListAPIKeys(c *gin.Context) {
	var keys []model.APIKey
	h.db.Find(&keys)
	c.JSON(http.StatusOK, keys)
}

func (h *AdminHandler) CreateAPIKey(c *gin.Context) {
	var input struct {
		Name          string   `json:"name"`
		AllowedModels []string `json:"allowed_models"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}

	keyStr, err := generateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errResp("failed to generate key"))
		return
	}

	allowedJSON := "[]"
	if len(input.AllowedModels) > 0 {
		b, _ := json.Marshal(input.AllowedModels)
		allowedJSON = string(b)
	}

	k := model.APIKey{
		Key:           keyStr,
		Name:          input.Name,
		AllowedModels: allowedJSON,
		Enabled:       true,
	}
	if err := h.db.Create(&k).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.JSON(http.StatusCreated, k)
}

func (h *AdminHandler) UpdateAPIKey(c *gin.Context) {
	var k model.APIKey
	if err := h.db.First(&k, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, errResp("key not found"))
		return
	}
	var input struct {
		Name          string   `json:"name"`
		AllowedModels []string `json:"allowed_models"`
		Enabled       *bool    `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, errResp(err.Error()))
		return
	}
	if input.Name != "" {
		k.Name = input.Name
	}
	if input.AllowedModels != nil {
		b, _ := json.Marshal(input.AllowedModels)
		k.AllowedModels = string(b)
	}
	if input.Enabled != nil {
		k.Enabled = *input.Enabled
	}
	h.db.Save(&k)
	c.JSON(http.StatusOK, k)
}

func (h *AdminHandler) DeleteAPIKey(c *gin.Context) {
	if err := h.db.Delete(&model.APIKey{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, errResp(err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Usage ---

func (h *AdminHandler) GetUsage(c *gin.Context) {
	date := c.Query("date")    // YYYY-MM-DD, optional
	keyID := c.Query("key_id") // optional

	query := h.db.Model(&model.UsageLog{})
	if date != "" {
		query = query.Where("date = ?", date)
	}
	if keyID != "" {
		query = query.Where("api_key_id = ?", keyID)
	}

	var logs []model.UsageLog
	query.Order("date DESC, total_tokens DESC").Find(&logs)
	c.JSON(http.StatusOK, logs)
}

// --- helpers ---

func generateKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "zapi-" + hex.EncodeToString(b), nil
}
