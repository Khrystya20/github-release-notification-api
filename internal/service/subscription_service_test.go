package service

import (
	"errors"
	"github-release-notification-api/internal/token"
	"testing"

	"github-release-notification-api/internal/model"
)

type fakeRepoRepo struct {
	getByFullNameResult *model.Repository
	getByFullNameErr    error

	createCalled bool
	createInput  *model.Repository
	createResult *model.Repository
	createErr    error

	getByIDResult *model.Repository
	getByIDErr    error
}

func (f *fakeRepoRepo) GetByFullName(fullName string) (*model.Repository, error) {
	if f.getByFullNameErr != nil {
		return nil, f.getByFullNameErr
	}
	if f.getByFullNameResult != nil {
		return f.getByFullNameResult, nil
	}
	return nil, nil
}

func (f *fakeRepoRepo) Create(repo *model.Repository) (*model.Repository, error) {
	f.createCalled = true
	f.createInput = repo

	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createResult != nil {
		return f.createResult, nil
	}

	repo.ID = 1
	return repo, nil
}

func (f *fakeRepoRepo) GetByID(id int64) (*model.Repository, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	if f.getByIDResult != nil {
		return f.getByIDResult, nil
	}
	return &model.Repository{
		ID:       id,
		FullName: "golang/go",
		Owner:    "golang",
		Name:     "go",
	}, nil
}

func (f *fakeRepoRepo) UpdateLastSeenTag(repositoryID int64, tag *string) error {
	return nil
}

func (f *fakeRepoRepo) GetTrackedRepositories() ([]model.Repository, error) {
	return nil, nil
}

func (f *fakeRepoRepo) UpdateLastCheckedAt(repositoryID int64) error {
	return nil
}

type fakeSubRepo struct {
	getByEmailAndRepositoryIDResult *model.Subscription
	getByEmailAndRepositoryIDErr    error

	createCalled bool
	createInput  *model.Subscription
	createResult *model.Subscription
	createErr    error

	getByConfirmTokenResult *model.Subscription
	getByConfirmTokenErr    error

	confirmByIDCalled bool
	confirmByIDValue  int64
	confirmByIDErr    error

	getByUnsubscribeTokenResult *model.Subscription
	getByUnsubscribeTokenErr    error

	deactivateByIDCalled bool
	deactivateByIDValue  int64
	deactivateByIDErr    error

	getActiveByEmailResult []model.SubscriptionResponse
	getActiveByEmailErr    error
	getActiveByEmailValue  string

	reactivateByIDCalled           bool
	reactivateByIDValue            int64
	reactivateByIDConfirmToken     string
	reactivateByIDUnsubscribeToken string
	reactivateByIDErr              error
}

func (f *fakeSubRepo) GetByEmailAndRepositoryID(email string, repositoryID int64) (*model.Subscription, error) {
	if f.getByEmailAndRepositoryIDErr != nil {
		return nil, f.getByEmailAndRepositoryIDErr
	}
	return f.getByEmailAndRepositoryIDResult, nil
}

func (f *fakeSubRepo) Create(subscription *model.Subscription) (*model.Subscription, error) {
	f.createCalled = true
	f.createInput = subscription

	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createResult != nil {
		return f.createResult, nil
	}

	subscription.ID = 1
	return subscription, nil
}

func (f *fakeSubRepo) GetActiveByEmail(email string) ([]model.SubscriptionResponse, error) {
	f.getActiveByEmailValue = email

	if f.getActiveByEmailErr != nil {
		return nil, f.getActiveByEmailErr
	}
	return f.getActiveByEmailResult, nil
}

func (f *fakeSubRepo) GetByConfirmToken(token string) (*model.Subscription, error) {
	if f.getByConfirmTokenErr != nil {
		return nil, f.getByConfirmTokenErr
	}
	return f.getByConfirmTokenResult, nil
}

func (f *fakeSubRepo) GetByUnsubscribeToken(token string) (*model.Subscription, error) {
	if f.getByUnsubscribeTokenErr != nil {
		return nil, f.getByUnsubscribeTokenErr
	}
	return f.getByUnsubscribeTokenResult, nil
}

func (f *fakeSubRepo) DeactivateByID(id int64) error {
	f.deactivateByIDCalled = true
	f.deactivateByIDValue = id
	return f.deactivateByIDErr
}

func (f *fakeSubRepo) ReactivateByID(id int64, confirmToken, unsubscribeToken string) error {
	f.reactivateByIDCalled = true
	f.reactivateByIDValue = id
	f.reactivateByIDConfirmToken = confirmToken
	f.reactivateByIDUnsubscribeToken = unsubscribeToken
	return f.reactivateByIDErr
}

func (f *fakeSubRepo) ConfirmByID(id int64) error {
	f.confirmByIDCalled = true
	f.confirmByIDValue = id
	return f.confirmByIDErr
}

func (f *fakeSubRepo) GetActiveConfirmedByRepositoryID(repositoryID int64) ([]model.Subscription, error) {
	return nil, nil
}

type fakeGitHub struct {
	repositoryExistsResult bool
	repositoryExistsErr    error

	getLatestReleaseTagResult *string
	getLatestReleaseTagErr    error
}

func (f *fakeGitHub) RepositoryExists(owner, repo string) (bool, error) {
	if f.repositoryExistsErr != nil {
		return false, f.repositoryExistsErr
	}
	return f.repositoryExistsResult, nil
}

func (f *fakeGitHub) GetLatestReleaseTag(owner, repo string) (*string, error) {
	if f.getLatestReleaseTagErr != nil {
		return nil, f.getLatestReleaseTagErr
	}
	return f.getLatestReleaseTagResult, nil
}

type fakeMailer struct {
	sendConfirmationEmailCalled bool
	sendConfirmationEmailTo     string
	sendConfirmationEmailRepo   string
	sendConfirmationEmailErr    error

	sendNewReleaseEmailCalled           bool
	sendNewReleaseEmailTo               string
	sendNewReleaseEmailRepo             string
	sendNewReleaseEmailTag              string
	sendNewReleaseEmailUnsubscribeToken string
	sendNewReleaseEmailErr              error
}

func (f *fakeMailer) SendConfirmationEmail(email, confirmToken, unsubscribeToken, repo string) error {
	f.sendConfirmationEmailCalled = true
	f.sendConfirmationEmailTo = email
	f.sendConfirmationEmailRepo = repo
	return f.sendConfirmationEmailErr
}

func (f *fakeMailer) SendNewReleaseEmail(email, repo, tag, unsubscribeToken string) error {
	f.sendNewReleaseEmailCalled = true
	f.sendNewReleaseEmailTo = email
	f.sendNewReleaseEmailRepo = repo
	f.sendNewReleaseEmailTag = tag
	f.sendNewReleaseEmailUnsubscribeToken = unsubscribeToken
	return f.sendNewReleaseEmailErr
}

func TestSubscribe_Success(t *testing.T) {
	tag := "v1.0.0"

	repoRepo := &fakeRepoRepo{
		getByFullNameResult: nil,
	}

	subRepo := &fakeSubRepo{}

	github := &fakeGitHub{
		repositoryExistsResult:    true,
		getLatestReleaseTagResult: &tag,
	}

	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !repoRepo.createCalled {
		t.Fatal("expected repository to be created")
	}

	if repoRepo.createInput == nil {
		t.Fatal("expected repository input to be saved")
	}

	if repoRepo.createInput.FullName != "golang/go" {
		t.Fatalf("expected repository full name golang/go, got %s", repoRepo.createInput.FullName)
	}

	if repoRepo.createInput.Owner != "golang" {
		t.Fatalf("expected repository owner golang, got %s", repoRepo.createInput.Owner)
	}

	if repoRepo.createInput.Name != "go" {
		t.Fatalf("expected repository name go, got %s", repoRepo.createInput.Name)
	}

	if repoRepo.createInput.LastSeenTag == nil {
		t.Fatal("expected repository last seen tag to be saved")
	}

	if *repoRepo.createInput.LastSeenTag != tag {
		t.Fatalf("expected last seen tag to be %s, got %s", tag, *repoRepo.createInput.LastSeenTag)
	}

	if !subRepo.createCalled {
		t.Fatal("expected subscription to be created")
	}

	if subRepo.createInput == nil {
		t.Fatal("expected created subscription input to be saved")
	}

	if subRepo.createInput.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", subRepo.createInput.Email)
	}

	if !mailer.sendConfirmationEmailCalled {
		t.Fatal("expected confirmation email to be sent")
	}

	if mailer.sendConfirmationEmailTo != "test@example.com" {
		t.Fatalf("expected email sent to test@example.com, got %s", mailer.sendConfirmationEmailTo)
	}

	if mailer.sendConfirmationEmailRepo != "golang/go" {
		t.Fatalf("expected repo golang/go, got %s", mailer.sendConfirmationEmailRepo)
	}
}

func TestSubscribe_InvalidEmail(t *testing.T) {
	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		&fakeGitHub{},
		&fakeMailer{},
	)

	err := service.Subscribe("invalid-email", "golang/go")

	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestSubscribe_RepoNotFound(t *testing.T) {
	github := &fakeGitHub{
		repositoryExistsResult: false,
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		github,
		&fakeMailer{},
	)

	err := service.Subscribe("test@example.com", "golang/go")

	if !errors.Is(err, ErrRepositoryNotFound) {
		t.Fatalf("expected ErrRepositoryNotFound, got %v", err)
	}
}

func TestSubscribe_CreateSubscriptionFails(t *testing.T) {
	tag := "v1.0.0"

	repoRepo := &fakeRepoRepo{}
	subRepo := &fakeSubRepo{
		createErr: errors.New("db error"),
	}
	github := &fakeGitHub{
		repositoryExistsResult:    true,
		getLatestReleaseTagResult: &tag,
	}
	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !subRepo.createCalled {
		t.Fatal("expected subscription create to be attempted")
	}

	if mailer.sendConfirmationEmailCalled {
		t.Fatal("email should not be sent when subscription creation fails")
	}
}

func mustGenerateToken(t *testing.T) string {
	t.Helper()

	value, err := token.Generate()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	return value
}

func TestConfirm_Success(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByConfirmTokenResult: &model.Subscription{
			ID:           1,
			RepositoryID: 1,
			Confirmed:    false,
			Active:       true,
		},
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Confirm(validToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !subRepo.confirmByIDCalled {
		t.Fatal("expected ConfirmByID to be called")
	}

	if subRepo.confirmByIDValue != 1 {
		t.Fatalf("expected ConfirmByID to be called with 1, got %d", subRepo.confirmByIDValue)
	}
}

func TestConfirm_TokenNotFound(t *testing.T) {
	validToken, err := token.Generate()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	subRepo := &fakeSubRepo{
		getByConfirmTokenResult: nil,
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	err = service.Confirm(validToken)
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestConfirm_EmptyToken(t *testing.T) {
	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		&fakeGitHub{},
		&fakeMailer{},
	)

	err := service.Confirm("")

	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestConfirm_AlreadyConfirmed(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByConfirmTokenResult: &model.Subscription{
			ID:           1,
			RepositoryID: 1,
			Confirmed:    true,
			Active:       true,
		},
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Confirm(validToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if subRepo.confirmByIDCalled {
		t.Fatal("ConfirmByID should not be called for already confirmed subscription")
	}
}

func TestConfirm_ConfirmByIDFails(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByConfirmTokenResult: &model.Subscription{
			ID:        1,
			Confirmed: false,
			Active:    true,
		},
		confirmByIDErr: errors.New("db update failed"),
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Confirm(validToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConfirm_InactiveSubscription(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByConfirmTokenResult: &model.Subscription{
			ID:        1,
			Confirmed: false,
			Active:    false,
		},
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Confirm(validToken)

	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}

	if subRepo.confirmByIDCalled {
		t.Fatal("ConfirmByID should not be called for inactive subscription")
	}
}

func TestSubscribe_InvalidRepoFormat(t *testing.T) {
	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		&fakeGitHub{},
		&fakeMailer{},
	)

	err := service.Subscribe("test@example.com", "invalid-repo-format")

	if !errors.Is(err, ErrInvalidRepoFormat) {
		t.Fatalf("expected ErrInvalidRepoFormat, got %v", err)
	}
}

func TestSubscribe_AlreadySubscribed(t *testing.T) {
	tag := "v1.0.0"

	repoRepo := &fakeRepoRepo{
		getByFullNameResult: &model.Repository{
			ID:       1,
			FullName: "golang/go",
			Owner:    "golang",
			Name:     "go",
		},
	}

	subRepo := &fakeSubRepo{
		getByEmailAndRepositoryIDResult: &model.Subscription{
			ID:           10,
			Email:        "test@example.com",
			RepositoryID: 1,
			Confirmed:    true,
			Active:       true,
		},
	}

	github := &fakeGitHub{
		repositoryExistsResult:    true,
		getLatestReleaseTagResult: &tag,
	}

	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")

	if !errors.Is(err, ErrAlreadySubscribed) {
		t.Fatalf("expected ErrAlreadySubscribed, got %v", err)
	}

	if subRepo.createCalled {
		t.Fatal("did not expect subscription to be created when it already exists")
	}

	if mailer.sendConfirmationEmailCalled {
		t.Fatal("did not expect confirmation email to be sent when subscription already exists")
	}
}

func TestSubscribe_ResendConfirmationForActiveUnconfirmedSubscription(t *testing.T) {
	repoRepo := &fakeRepoRepo{
		getByFullNameResult: &model.Repository{
			ID:       1,
			FullName: "golang/go",
			Owner:    "golang",
			Name:     "go",
		},
	}

	subRepo := &fakeSubRepo{
		getByEmailAndRepositoryIDResult: &model.Subscription{
			ID:           10,
			Email:        "test@example.com",
			RepositoryID: 1,
			Confirmed:    false,
			Active:       true,
		},
	}

	github := &fakeGitHub{
		repositoryExistsResult: true,
	}

	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if subRepo.createCalled {
		t.Fatal("did not expect new subscription to be created")
	}

	if !subRepo.reactivateByIDCalled {
		t.Fatal("expected ReactivateByID to be called")
	}

	if subRepo.reactivateByIDValue != 10 {
		t.Fatalf("expected ReactivateByID to be called with 10, got %d", subRepo.reactivateByIDValue)
	}

	if subRepo.reactivateByIDConfirmToken == "" {
		t.Fatal("expected confirm token to be passed to ReactivateByID")
	}

	if subRepo.reactivateByIDUnsubscribeToken == "" {
		t.Fatal("expected unsubscribe token to be passed to ReactivateByID")
	}

	if !mailer.sendConfirmationEmailCalled {
		t.Fatal("expected confirmation email to be sent")
	}
}

func TestSubscribe_ReactivateInactiveSubscription(t *testing.T) {
	repoRepo := &fakeRepoRepo{
		getByFullNameResult: &model.Repository{
			ID:       1,
			FullName: "golang/go",
			Owner:    "golang",
			Name:     "go",
		},
	}

	subRepo := &fakeSubRepo{
		getByEmailAndRepositoryIDResult: &model.Subscription{
			ID:           10,
			Email:        "test@example.com",
			RepositoryID: 1,
			Confirmed:    false,
			Active:       false,
		},
	}

	github := &fakeGitHub{
		repositoryExistsResult: true,
	}

	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if subRepo.createCalled {
		t.Fatal("did not expect new subscription to be created")
	}

	if !subRepo.reactivateByIDCalled {
		t.Fatal("expected ReactivateByID to be called")
	}

	if subRepo.reactivateByIDValue != 10 {
		t.Fatalf("expected ReactivateByID to be called with 10, got %d", subRepo.reactivateByIDValue)
	}

	if subRepo.reactivateByIDConfirmToken == "" {
		t.Fatal("expected confirm token to be passed to ReactivateByID")
	}

	if subRepo.reactivateByIDUnsubscribeToken == "" {
		t.Fatal("expected unsubscribe token to be passed to ReactivateByID")
	}

	if !mailer.sendConfirmationEmailCalled {
		t.Fatal("expected confirmation email to be sent")
	}
}

func TestSubscribe_ReactivateInactiveSubscriptionFails(t *testing.T) {
	repoRepo := &fakeRepoRepo{
		getByFullNameResult: &model.Repository{
			ID:       1,
			FullName: "golang/go",
			Owner:    "golang",
			Name:     "go",
		},
	}

	subRepo := &fakeSubRepo{
		getByEmailAndRepositoryIDResult: &model.Subscription{
			ID:           10,
			Email:        "test@example.com",
			RepositoryID: 1,
			Confirmed:    false,
			Active:       false,
		},
		reactivateByIDErr: errors.New("db update failed"),
	}

	github := &fakeGitHub{
		repositoryExistsResult: true,
	}

	mailer := &fakeMailer{}

	service := NewSubscriptionService(repoRepo, subRepo, github, mailer)

	err := service.Subscribe("test@example.com", "golang/go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !subRepo.reactivateByIDCalled {
		t.Fatal("expected ReactivateByID to be called")
	}

	if mailer.sendConfirmationEmailCalled {
		t.Fatal("email should not be sent when reactivation fails")
	}
}

func TestUnsubscribe_EmptyToken(t *testing.T) {
	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		&fakeGitHub{},
		&fakeMailer{},
	)

	err := service.Unsubscribe("")

	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestUnsubscribe_TokenNotFound(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByUnsubscribeTokenResult: nil,
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Unsubscribe(validToken)

	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}

func TestUnsubscribe_Success(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByUnsubscribeTokenResult: &model.Subscription{
			ID: 5,
		},
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Unsubscribe(validToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !subRepo.deactivateByIDCalled {
		t.Fatal("expected DeactivateByID to be called")
	}

	if subRepo.deactivateByIDValue != 5 {
		t.Fatalf("expected DeactivateByID to be called with 5, got %d", subRepo.deactivateByIDValue)
	}
}

func TestUnsubscribe_DeactivateFails(t *testing.T) {
	subRepo := &fakeSubRepo{
		getByUnsubscribeTokenResult: &model.Subscription{
			ID: 5,
		},
		deactivateByIDErr: errors.New("db error"),
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	validToken := mustGenerateToken(t)
	err := service.Unsubscribe(validToken)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSubscriptions_InvalidEmail(t *testing.T) {
	service := NewSubscriptionService(
		&fakeRepoRepo{},
		&fakeSubRepo{},
		&fakeGitHub{},
		&fakeMailer{},
	)

	_, err := service.GetSubscriptions("bad-email")

	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestGetSubscriptions_Success(t *testing.T) {
	tag := "v1.0.0"

	subRepo := &fakeSubRepo{
		getActiveByEmailResult: []model.SubscriptionResponse{
			{
				Email:       "test@example.com",
				Repo:        "golang/go",
				Confirmed:   true,
				LastSeenTag: &tag,
			},
		},
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	result, err := service.GetSubscriptions("test@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if subRepo.getActiveByEmailValue != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", subRepo.getActiveByEmailValue)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(result))
	}

	if result[0].Repo != "golang/go" {
		t.Fatalf("expected repo golang/go, got %s", result[0].Repo)
	}
}

func TestGetSubscriptions_RepoFails(t *testing.T) {
	subRepo := &fakeSubRepo{
		getActiveByEmailErr: errors.New("db error"),
	}

	service := NewSubscriptionService(
		&fakeRepoRepo{},
		subRepo,
		&fakeGitHub{},
		&fakeMailer{},
	)

	_, err := service.GetSubscriptions("test@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
