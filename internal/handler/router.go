package handler

import "github.com/gin-gonic/gin"

func SetupRouter(subscriptionHandler *SubscriptionHandler) *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscriptionHandler.Subscribe)
		api.GET("/confirm/:token", subscriptionHandler.Confirm)
		api.GET("/unsubscribe/:token", subscriptionHandler.Unsubscribe)
		api.GET("/subscriptions", subscriptionHandler.GetSubscriptions)
	}

	return router
}
