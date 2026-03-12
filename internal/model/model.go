package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Provider represents an upstream LLM API provider.
type Provider struct {
	ID      uint           `gorm:"primaryKey"          json:"id"`
	Name    string         `gorm:"uniqueIndex;not null" json:"name"`
	BaseURL string         `gorm:"not null"             json:"base_url"`
	APIKey  string         `gorm:"not null"             json:"api_key"`
	// APIType: "openai" (default) or "anthropic"
	APIType   string         `gorm:"default:openai"       json:"api_type"`
	Enabled   bool           `gorm:"default:true"         json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                json:"-"`
	Models    []ProviderModel `gorm:"foreignKey:ProviderID" json:"models,omitempty"`
}

// ProviderModel is a model offered by a Provider.
type ProviderModel struct {
	ID         uint           `gorm:"primaryKey"     json:"id"`
	ProviderID uint           `gorm:"not null;index" json:"provider_id"`
	// ModelID is the identifier exposed to API consumers.
	ModelID string `gorm:"not null" json:"model_id"`
	// ProviderModelID is the actual model ID sent to the upstream (defaults to ModelID if empty).
	ProviderModelID string         `json:"provider_model_id"`
	Alias           string         `json:"alias"`
	Enabled         bool           `gorm:"default:true" json:"enabled"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index"        json:"-"`
	Provider        Provider       `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
}

// UpstreamModelID returns the model ID to send to the upstream provider.
func (m *ProviderModel) UpstreamModelID() string {
	if m.ProviderModelID != "" {
		return m.ProviderModelID
	}
	return m.ModelID
}

// APIKey is a consumer key that authorizes access to zAPI.
type APIKey struct {
	ID   uint   `gorm:"primaryKey"           json:"id"`
	Key  string `gorm:"uniqueIndex;not null" json:"key"`
	Name string `json:"name"`
	// AllowedModels is a JSON array of model IDs; empty means all models are allowed.
	AllowedModels string         `gorm:"type:text"            json:"allowed_models"`
	Enabled       bool           `gorm:"default:true"         json:"enabled"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                json:"-"`
}

// HasModelAccess reports whether this key may use the given model.
func (k *APIKey) HasModelAccess(modelID string) bool {
	if k.AllowedModels == "" || k.AllowedModels == "[]" {
		return true
	}
	var allowed []string
	if err := json.Unmarshal([]byte(k.AllowedModels), &allowed); err != nil {
		return false
	}
	for _, m := range allowed {
		if m == modelID {
			return true
		}
	}
	return false
}

// UsageLog records token consumption per key/model/day.
type UsageLog struct {
	ID               uint      `gorm:"primaryKey"     json:"id"`
	APIKeyID         uint      `gorm:"not null;index" json:"api_key_id"`
	APIKeyName       string    `json:"api_key_name"`
	Model            string    `gorm:"not null"       json:"model"`
	Date             string    `gorm:"not null;index" json:"date"` // YYYY-MM-DD
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	RequestCount     int       `gorm:"default:1"      json:"request_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
