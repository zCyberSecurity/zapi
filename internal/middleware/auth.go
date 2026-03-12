package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zCyberSecurity/zapi/internal/model"
	"gorm.io/gorm"
)

const CtxAPIKey = "api_key"

func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := bearerToken(c)
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}

		var apiKey model.APIKey
		if err := db.Where("key = ? AND enabled = ?", key, true).First(&apiKey).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or disabled API key"})
			return
		}

		c.Set(CtxAPIKey, &apiKey)
		c.Next()
	}
}

func AdminAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if bearerToken(c) != token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid admin token"})
			return
		}
		c.Next()
	}
}

func bearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
