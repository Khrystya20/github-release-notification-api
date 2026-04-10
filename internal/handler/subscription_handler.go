package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github-release-notification-api/internal/model"
	"github-release-notification-api/internal/service"
)

type SubscriptionService interface {
	Subscribe(email, repo string) error
	Confirm(token string) error
	Unsubscribe(token string) error
	GetSubscriptions(email string) ([]model.SubscriptionResponse, error)
}

type SubscriptionHandler struct {
	service SubscriptionService
}

func NewSubscriptionHandler(service SubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

type SubscribeRequest struct {
	Email string `json:"email"`
	Repo  string `json:"repo"`
}

func (h *SubscriptionHandler) Subscribe(c *gin.Context) {
	var req SubscribeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.service.Subscribe(req.Email, req.Repo)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail),
			errors.Is(err, service.ErrInvalidRepoFormat):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		case errors.Is(err, service.ErrRepositoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return

		case errors.Is(err, service.ErrAlreadySubscribed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return

		case errors.Is(err, service.ErrGitHubRateLimited):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "github api rate limit exceeded, please try again later",
			})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "request cannot be processed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription successful. Confirmation email sent.",
	})
}

func (h *SubscriptionHandler) Confirm(c *gin.Context) {
	token := c.Param("token")

	err := h.service.Confirm(token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		case errors.Is(err, service.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "request cannot be processed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription confirmed successfully",
	})
}

func (h *SubscriptionHandler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")

	err := h.service.Unsubscribe(token)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		case errors.Is(err, service.ErrTokenNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "request cannot be processed"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Unsubscribed successfully",
	})
}

func (h *SubscriptionHandler) GetSubscriptions(c *gin.Context) {
	email := c.Query("email")

	subscriptions, err := h.service.GetSubscriptions(email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "request cannot be processed"})
			return
		}
	}

	c.JSON(http.StatusOK, subscriptions)
}
