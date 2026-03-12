package router

import (
	"github.com/gin-gonic/gin"
	"github.com/zCyberSecurity/zapi/internal/config"
	"github.com/zCyberSecurity/zapi/internal/handler"
	"github.com/zCyberSecurity/zapi/internal/middleware"
	"gorm.io/gorm"
)

func New(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	oaiH := handler.NewOpenAIHandler(db)
	anthH := handler.NewAnthropicHandler(db)
	admH := handler.NewAdminHandler(db)

	// OpenAI-compatible endpoints
	v1 := r.Group("/v1", middleware.APIKeyAuth(db))
	{
		v1.POST("/chat/completions", oaiH.ChatCompletions)
		v1.GET("/models", oaiH.ListModels)
	}

	// Anthropic-compatible endpoint
	r.POST("/v1/messages", middleware.APIKeyAuth(db), anthH.CreateMessage)

	// Admin endpoints
	adm := r.Group("/admin", middleware.AdminAuth(cfg.AdminToken))
	{
		adm.GET("/providers", admH.ListProviders)
		adm.POST("/providers", admH.CreateProvider)
		adm.PUT("/providers/:id", admH.UpdateProvider)
		adm.DELETE("/providers/:id", admH.DeleteProvider)

		adm.GET("/providers/:id/models", admH.ListModels)
		adm.POST("/providers/:id/models", admH.AddModel)
		adm.PUT("/models/:id", admH.UpdateModel)
		adm.DELETE("/models/:id", admH.DeleteModel)

		adm.GET("/keys", admH.ListAPIKeys)
		adm.POST("/keys", admH.CreateAPIKey)
		adm.PUT("/keys/:id", admH.UpdateAPIKey)
		adm.DELETE("/keys/:id", admH.DeleteAPIKey)

		adm.GET("/usage", admH.GetUsage)
	}

	return r
}
