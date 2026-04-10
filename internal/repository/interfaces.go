package repository

import "github-release-notification-api/internal/model"

type RepositoryRepository interface {
	GetByFullName(fullName string) (*model.Repository, error)
	GetByID(id int64) (*model.Repository, error)
	Create(repo *model.Repository) (*model.Repository, error)
	UpdateLastSeenTag(repositoryID int64, tag *string) error
	GetTrackedRepositories() ([]model.Repository, error)
	UpdateLastCheckedAt(repositoryID int64) error
}

type SubscriptionRepository interface {
	GetByEmailAndRepositoryID(email string, repositoryID int64) (*model.Subscription, error)
	Create(subscription *model.Subscription) (*model.Subscription, error)
	GetActiveByEmail(email string) ([]model.SubscriptionResponse, error)
	GetByConfirmToken(token string) (*model.Subscription, error)
	GetByUnsubscribeToken(token string) (*model.Subscription, error)
	ConfirmByID(id int64) error
	DeactivateByID(id int64) error
	ReactivateByID(id int64, confirmToken, unsubscribeToken string) error
	GetActiveConfirmedByRepositoryID(repositoryID int64) ([]model.Subscription, error)
}
