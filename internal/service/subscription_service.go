package service

import (
	"github-release-notification-api/internal/token"
	"strings"

	"github-release-notification-api/internal/model"
	"github-release-notification-api/internal/repository"
	"github-release-notification-api/internal/validator"
)

type SubscriptionService struct {
	repoRepo         repository.RepositoryRepository
	subscriptionRepo repository.SubscriptionRepository
	githubClient     GitHubClient
	mailSender       MailSender
}

func NewSubscriptionService(
	repoRepo repository.RepositoryRepository,
	subscriptionRepo repository.SubscriptionRepository,
	githubClient GitHubClient,
	mailSender MailSender,
) *SubscriptionService {
	return &SubscriptionService{
		repoRepo:         repoRepo,
		subscriptionRepo: subscriptionRepo,
		githubClient:     githubClient,
		mailSender:       mailSender,
	}
}

func (s *SubscriptionService) Subscribe(email, repo string) error {
	email = strings.TrimSpace(email)
	repo = strings.TrimSpace(repo)

	if err := validator.ValidateEmail(email); err != nil {
		return ErrInvalidEmail
	}

	owner, name, err := validator.ParseAndValidateRepo(repo)
	if err != nil {
		return ErrInvalidRepoFormat
	}

	exists, err := s.githubClient.RepositoryExists(owner, name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "rate limit") {
			return ErrGitHubRateLimited
		}
		return err
	}
	if !exists {
		return ErrRepositoryNotFound
	}

	fullName := owner + "/" + name

	repositoryModel, err := s.repoRepo.GetByFullName(fullName)
	if err != nil {
		return err
	}

	if repositoryModel == nil {
		lastSeenTag, err := s.getInitialLastSeenTag(owner, name)
		if err != nil {
			return err
		}

		repositoryModel = &model.Repository{
			FullName:    fullName,
			Owner:       owner,
			Name:        name,
			LastSeenTag: lastSeenTag,
		}

		repositoryModel, err = s.repoRepo.Create(repositoryModel)
		if err != nil {
			return err
		}
	}

	existingSubscription, err := s.subscriptionRepo.GetByEmailAndRepositoryID(email, repositoryModel.ID)
	if err != nil {
		return err
	}

	confirmToken, err := token.Generate()
	if err != nil {
		return err
	}

	unsubscribeToken, err := token.Generate()
	if err != nil {
		return err
	}

	if existingSubscription != nil {
		if existingSubscription.Active && existingSubscription.Confirmed {
			return ErrAlreadySubscribed
		}

		err = s.subscriptionRepo.ReactivateByID(
			existingSubscription.ID,
			confirmToken,
			unsubscribeToken,
		)
		if err != nil {
			return err
		}

		err = s.mailSender.SendConfirmationEmail(email, confirmToken, unsubscribeToken, fullName)
		if err != nil {
			return err
		}

		return nil
	}

	subscription := &model.Subscription{
		Email:            email,
		RepositoryID:     repositoryModel.ID,
		Confirmed:        false,
		Active:           true,
		ConfirmToken:     confirmToken,
		UnsubscribeToken: unsubscribeToken,
	}

	_, err = s.subscriptionRepo.Create(subscription)
	if err != nil {
		return err
	}

	err = s.mailSender.SendConfirmationEmail(email, confirmToken, unsubscribeToken, fullName)
	if err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) getInitialLastSeenTag(owner, name string) (*string, error) {
	tag, err := s.githubClient.GetLatestReleaseTag(owner, name)
	if err != nil {
		errText := strings.ToLower(err.Error())

		if strings.Contains(errText, "rate limit") {
			return nil, ErrGitHubRateLimited
		}

		return nil, err
	}

	if tag == nil {
		return nil, nil
	}

	if strings.TrimSpace(*tag) == "" {
		return nil, nil
	}

	return tag, nil
}

func (s *SubscriptionService) Confirm(token string) error {
	if err := validator.ValidateToken(token); err != nil {
		return ErrInvalidToken
	}

	subscription, err := s.subscriptionRepo.GetByConfirmToken(token)
	if err != nil {
		return err
	}
	if subscription == nil {
		return ErrTokenNotFound
	}

	if !subscription.Active {
		return ErrTokenNotFound
	}

	if subscription.Confirmed {
		return nil
	}

	if err := s.subscriptionRepo.ConfirmByID(subscription.ID); err != nil {
		return err
	}

	return nil
}

func (s *SubscriptionService) Unsubscribe(token string) error {
	if err := validator.ValidateToken(token); err != nil {
		return ErrInvalidToken
	}

	subscription, err := s.subscriptionRepo.GetByUnsubscribeToken(token)
	if err != nil {
		return err
	}
	if subscription == nil {
		return ErrTokenNotFound
	}

	return s.subscriptionRepo.DeactivateByID(subscription.ID)
}

func (s *SubscriptionService) GetSubscriptions(email string) ([]model.SubscriptionResponse, error) {
	email = strings.TrimSpace(email)

	if err := validator.ValidateEmail(email); err != nil {
		return nil, ErrInvalidEmail
	}

	return s.subscriptionRepo.GetActiveByEmail(email)
}
