package service

import (
	"errors"
	"testing"

	"github-release-notification-api/internal/model"
)

type fakeScannerRepoRepo struct {
	getTrackedRepositoriesResult []model.Repository
	getTrackedRepositoriesErr    error

	updateLastSeenTagCalled bool
	updateLastSeenTagRepoID int64
	updateLastSeenTagValue  *string
	updateLastSeenTagErr    error

	updateLastCheckedAtCalled bool
	updateLastCheckedAtRepoID int64
	updateLastCheckedAtErr    error
}

func (f *fakeScannerRepoRepo) GetByFullName(fullName string) (*model.Repository, error) {
	return nil, nil
}

func (f *fakeScannerRepoRepo) Create(repo *model.Repository) (*model.Repository, error) {
	return nil, nil
}

func (f *fakeScannerRepoRepo) GetByID(id int64) (*model.Repository, error) {
	return nil, nil
}

func (f *fakeScannerRepoRepo) UpdateLastSeenTag(repositoryID int64, tag *string) error {
	f.updateLastSeenTagCalled = true
	f.updateLastSeenTagRepoID = repositoryID
	f.updateLastSeenTagValue = tag
	return f.updateLastSeenTagErr
}

func (f *fakeScannerRepoRepo) GetTrackedRepositories() ([]model.Repository, error) {
	if f.getTrackedRepositoriesErr != nil {
		return nil, f.getTrackedRepositoriesErr
	}
	return f.getTrackedRepositoriesResult, nil
}

func (f *fakeScannerRepoRepo) UpdateLastCheckedAt(repositoryID int64) error {
	f.updateLastCheckedAtCalled = true
	f.updateLastCheckedAtRepoID = repositoryID
	return f.updateLastCheckedAtErr
}

type fakeScannerSubRepo struct {
	getActiveConfirmedByRepositoryIDResult []model.Subscription
	getActiveConfirmedByRepositoryIDErr    error

	requestedRepositoryID int64

	reactivateByIDCalled           bool
	reactivateByIDID               int64
	reactivateByIDConfirmToken     string
	reactivateByIDUnsubscribeToken string
	reactivateByIDErr              error
}

func (f *fakeScannerSubRepo) ReactivateByID(id int64, confirmToken, unsubscribeToken string) error {
	f.reactivateByIDCalled = true
	f.reactivateByIDID = id
	f.reactivateByIDConfirmToken = confirmToken
	f.reactivateByIDUnsubscribeToken = unsubscribeToken
	return f.reactivateByIDErr
}

func (f *fakeScannerSubRepo) GetByEmailAndRepositoryID(email string, repositoryID int64) (*model.Subscription, error) {
	return nil, nil
}

func (f *fakeScannerSubRepo) Create(subscription *model.Subscription) (*model.Subscription, error) {
	return nil, nil
}

func (f *fakeScannerSubRepo) GetActiveByEmail(email string) ([]model.SubscriptionResponse, error) {
	return nil, nil
}

func (f *fakeScannerSubRepo) GetByConfirmToken(token string) (*model.Subscription, error) {
	return nil, nil
}

func (f *fakeScannerSubRepo) GetByUnsubscribeToken(token string) (*model.Subscription, error) {
	return nil, nil
}

func (f *fakeScannerSubRepo) ConfirmByID(id int64) error {
	return nil
}

func (f *fakeScannerSubRepo) DeactivateByID(id int64) error {
	return nil
}

func (f *fakeScannerSubRepo) GetActiveConfirmedByRepositoryID(repositoryID int64) ([]model.Subscription, error) {
	f.requestedRepositoryID = repositoryID

	if f.getActiveConfirmedByRepositoryIDErr != nil {
		return nil, f.getActiveConfirmedByRepositoryIDErr
	}
	return f.getActiveConfirmedByRepositoryIDResult, nil
}

type fakeScannerGitHub struct {
	getLatestReleaseTagResult *string
	getLatestReleaseTagErr    error

	lastOwner string
	lastRepo  string
}

func (f *fakeScannerGitHub) RepositoryExists(owner, repo string) (bool, error) {
	return true, nil
}

func (f *fakeScannerGitHub) GetLatestReleaseTag(owner, repo string) (*string, error) {
	f.lastOwner = owner
	f.lastRepo = repo

	if f.getLatestReleaseTagErr != nil {
		return nil, f.getLatestReleaseTagErr
	}
	return f.getLatestReleaseTagResult, nil
}

type fakeScannerMailer struct {
	sendNewReleaseEmailCalled int
	sentTo                    []string
	sentRepo                  []string
	sentTag                   []string
	sentUnsubscribeToken      []string
	sendNewReleaseEmailErr    error
}

func (f *fakeScannerMailer) SendConfirmationEmail(email, confirmToken, unsubscribeToken, repo string) error {
	return nil
}

func (f *fakeScannerMailer) SendNewReleaseEmail(email, repo, tag, unsubscribeToken string) error {
	f.sendNewReleaseEmailCalled++
	f.sentTo = append(f.sentTo, email)
	f.sentRepo = append(f.sentRepo, repo)
	f.sentTag = append(f.sentTag, tag)
	f.sentUnsubscribeToken = append(f.sentUnsubscribeToken, unsubscribeToken)

	return f.sendNewReleaseEmailErr
}

func strPtr(s string) *string {
	return &s
}

func TestScanner_ScanOnce_NewRelease_SendsEmailsAndUpdatesRepository(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.0.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{
		getActiveConfirmedByRepositoryIDResult: []model.Subscription{
			{
				ID:               1,
				Email:            "user1@example.com",
				UnsubscribeToken: "token-1",
				Confirmed:        true,
			},
			{
				ID:               2,
				Email:            "user2@example.com",
				UnsubscribeToken: "token-2",
				Confirmed:        true,
			},
		},
	}

	github := &fakeScannerGitHub{
		getLatestReleaseTagResult: strPtr("v1.1.0"),
	}

	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if github.lastOwner != "golang" {
		t.Fatalf("expected owner golang, got %s", github.lastOwner)
	}

	if github.lastRepo != "go" {
		t.Fatalf("expected repo go, got %s", github.lastRepo)
	}

	if subRepo.requestedRepositoryID != 1 {
		t.Fatalf("expected subscriptions to be requested for repository 1, got %d", subRepo.requestedRepositoryID)
	}

	if mailer.sendNewReleaseEmailCalled != 2 {
		t.Fatalf("expected 2 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}

	if mailer.sentTo[0] != "user1@example.com" {
		t.Fatalf("expected first email to user1@example.com, got %s", mailer.sentTo[0])
	}

	if mailer.sentTo[1] != "user2@example.com" {
		t.Fatalf("expected second email to user2@example.com, got %s", mailer.sentTo[1])
	}

	if mailer.sentRepo[0] != "golang/go" {
		t.Fatalf("expected repo golang/go, got %s", mailer.sentRepo[0])
	}

	if mailer.sentTag[0] != "v1.1.0" {
		t.Fatalf("expected tag v1.1.0, got %s", mailer.sentTag[0])
	}

	if !repoRepo.updateLastSeenTagCalled {
		t.Fatal("expected UpdateLastSeenTag to be called")
	}

	if repoRepo.updateLastSeenTagRepoID != 1 {
		t.Fatalf("expected UpdateLastSeenTag for repository 1, got %d", repoRepo.updateLastSeenTagRepoID)
	}

	if repoRepo.updateLastSeenTagValue == nil || *repoRepo.updateLastSeenTagValue != "v1.1.0" {
		t.Fatal("expected last seen tag to be updated to v1.1.0")
	}

	if !repoRepo.updateLastCheckedAtCalled {
		t.Fatal("expected UpdateLastCheckedAt to be called")
	}

	if repoRepo.updateLastCheckedAtRepoID != 1 {
		t.Fatalf("expected UpdateLastCheckedAt for repository 1, got %d", repoRepo.updateLastCheckedAtRepoID)
	}
}

func TestScanner_ScanOnce_SameTag_DoesNotSendEmailsOrUpdateLastSeenTag(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.1.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{
		getActiveConfirmedByRepositoryIDResult: []model.Subscription{
			{
				ID:               1,
				Email:            "user@example.com",
				UnsubscribeToken: "token-1",
				Confirmed:        true,
			},
		},
	}

	github := &fakeScannerGitHub{
		getLatestReleaseTagResult: strPtr("v1.1.0"),
	}

	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if mailer.sendNewReleaseEmailCalled != 0 {
		t.Fatalf("expected 0 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}

	if subRepo.requestedRepositoryID != 0 {
		t.Fatalf("expected subscriptions not to be requested, got repository id %d", subRepo.requestedRepositoryID)
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called when tag did not change")
	}

	if !repoRepo.updateLastCheckedAtCalled {
		t.Fatal("expected UpdateLastCheckedAt to be called")
	}
}

func TestScanner_ScanOnce_GetTrackedRepositoriesFails_DoesNothing(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesErr: errors.New("db error"),
	}
	service := NewScannerService(repoRepo, &fakeScannerSubRepo{}, &fakeScannerGitHub{}, &fakeScannerMailer{})

	service.ScanOnce()

	if repoRepo.updateLastCheckedAtCalled {
		t.Fatal("did not expect UpdateLastCheckedAt to be called")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called")
	}
}

func TestScanner_ScanOnce_GetLatestReleaseTagFails_DoesNotUpdateRepository(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.0.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{}
	github := &fakeScannerGitHub{getLatestReleaseTagErr: errors.New("github error")}
	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if repoRepo.updateLastCheckedAtCalled {
		t.Fatal("did not expect UpdateLastCheckedAt to be called when GitHub request failed")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called when GitHub request failed")
	}

	if subRepo.requestedRepositoryID != 0 {
		t.Fatalf("expected subscriptions not to be requested, got repository id %d", subRepo.requestedRepositoryID)
	}

	if mailer.sendNewReleaseEmailCalled != 0 {
		t.Fatalf("expected 0 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}
}

func TestScanner_ScanOnce_RateLimitError_DoesNotUpdateRepository(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.0.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{}
	github := &fakeScannerGitHub{getLatestReleaseTagErr: errors.New("API rate limit exceeded")}
	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if repoRepo.updateLastCheckedAtCalled {
		t.Fatal("did not expect UpdateLastCheckedAt to be called on rate limit")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called on rate limit")
	}

	if mailer.sendNewReleaseEmailCalled != 0 {
		t.Fatalf("expected 0 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}
}

func TestScanner_ScanOnce_NoRelease_DoesNotSendEmailsButUpdatesLastCheckedAt(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:       1,
				FullName: "golang/go",
				Owner:    "golang",
				Name:     "go",
			},
		},
	}

	subRepo := &fakeScannerSubRepo{}
	github := &fakeScannerGitHub{getLatestReleaseTagResult: nil}
	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if !repoRepo.updateLastCheckedAtCalled {
		t.Fatal("expected UpdateLastCheckedAt to be called")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called when no release exists")
	}

	if subRepo.requestedRepositoryID != 0 {
		t.Fatalf("expected subscriptions not to be requested, got repository id %d", subRepo.requestedRepositoryID)
	}

	if mailer.sendNewReleaseEmailCalled != 0 {
		t.Fatalf("expected 0 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}
}

func TestScanner_ScanOnce_GetSubscribersFails_DoesNotUpdateLastSeenTag(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.0.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{
		getActiveConfirmedByRepositoryIDErr: errors.New("db error"),
	}

	github := &fakeScannerGitHub{
		getLatestReleaseTagResult: strPtr("v1.1.0"),
	}

	mailer := &fakeScannerMailer{}
	service := NewScannerService(repoRepo, subRepo, github, mailer)

	service.ScanOnce()

	if !repoRepo.updateLastCheckedAtCalled {
		t.Fatal("expected UpdateLastCheckedAt to be called")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called when loading subscriptions failed")
	}

	if mailer.sendNewReleaseEmailCalled != 0 {
		t.Fatalf("expected 0 emails to be sent, got %d", mailer.sendNewReleaseEmailCalled)
	}
}

func TestScanner_ScanOnce_SendEmailFails_DoesNotUpdateLastSeenTag(t *testing.T) {
	repoRepo := &fakeScannerRepoRepo{
		getTrackedRepositoriesResult: []model.Repository{
			{
				ID:          1,
				FullName:    "golang/go",
				Owner:       "golang",
				Name:        "go",
				LastSeenTag: strPtr("v1.0.0"),
			},
		},
	}

	subRepo := &fakeScannerSubRepo{
		getActiveConfirmedByRepositoryIDResult: []model.Subscription{
			{
				ID:               1,
				Email:            "user@example.com",
				UnsubscribeToken: "token-1",
				Confirmed:        true,
			},
		},
	}

	github := &fakeScannerGitHub{
		getLatestReleaseTagResult: strPtr("v1.1.0"),
	}

	mailer := &fakeScannerMailer{
		sendNewReleaseEmailErr: errors.New("smtp error"),
	}

	service := NewScannerService(repoRepo, subRepo, github, mailer)
	service.ScanOnce()

	if mailer.sendNewReleaseEmailCalled != 1 {
		t.Fatalf("expected 1 email attempt, got %d", mailer.sendNewReleaseEmailCalled)
	}

	if !repoRepo.updateLastCheckedAtCalled {
		t.Fatal("expected UpdateLastCheckedAt to be called")
	}

	if repoRepo.updateLastSeenTagCalled {
		t.Fatal("did not expect UpdateLastSeenTag to be called when email sending fails")
	}

}
