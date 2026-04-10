package handler

import (
	"embed"
	"html/template"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*.html
var templatesFS embed.FS

func SetupRouter(subscriptionHandler *SubscriptionHandler) *gin.Engine {
	router := gin.Default()

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/*.html"))
	router.SetHTMLTemplate(tmpl)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/subscribe")
	})

	router.GET("/subscribe", func(c *gin.Context) {
		c.HTML(200, "subscribe.html", nil)
	})

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionHandler.Subscribe)
		api.GET("/confirm/:token", subscriptionHandler.Confirm)
		api.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
		api.GET("/subscriptions", subscriptionHandler.GetSubscriptions)
	}

	return router
}
