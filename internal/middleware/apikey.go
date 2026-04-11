package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func APIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "api key is not configured",
			})
			c.Abort()
			return
		}

		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "missing api key",
			})
			c.Abort()
			return
		}

		if providedKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid api key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
