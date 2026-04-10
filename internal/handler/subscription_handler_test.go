package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github-release-notification-api/internal/model"
	"github-release-notification-api/internal/service"
)

type fakeSubscriptionService struct {
	subscribeErr   error
	confirmErr     error
	unsubscribeErr error

	getSubscriptionsResult []model.SubscriptionResponse
	getSubscriptionsErr    error

	subscribeCalled bool
	subscribeEmail  string
	subscribeRepo   string

	confirmCalled bool
	confirmToken  string

	unsubscribeCalled bool
	unsubscribeToken  string

	getSubscriptionsCalled bool
	getSubscriptionsEmail  string
}

func (f *fakeSubscriptionService) Subscribe(email, repo string) error {
	f.subscribeCalled = true
	f.subscribeEmail = email
	f.subscribeRepo = repo
	return f.subscribeErr
}

func (f *fakeSubscriptionService) Confirm(token string) error {
	f.confirmCalled = true
	f.confirmToken = token
	return f.confirmErr
}

func (f *fakeSubscriptionService) Unsubscribe(token string) error {
	f.unsubscribeCalled = true
	f.unsubscribeToken = token
	return f.unsubscribeErr
}

func (f *fakeSubscriptionService) GetSubscriptions(email string) ([]model.SubscriptionResponse, error) {
	f.getSubscriptionsCalled = true
	f.getSubscriptionsEmail = email
	return f.getSubscriptionsResult, f.getSubscriptionsErr
}

func setupTestRouter(svc SubscriptionService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewSubscriptionHandler(svc)
	return SetupRouter(h)
}

func newJSONRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func decodeMap(t *testing.T, w *httptest.ResponseRecorder) map[string]string {
	t.Helper()

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	return body
}

func TestSubscribe_Success(t *testing.T) {
	svc := &fakeSubscriptionService{}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"golang/go"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !svc.subscribeCalled {
		t.Fatal("expected Subscribe to be called")
	}

	if svc.subscribeEmail != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", svc.subscribeEmail)
	}

	if svc.subscribeRepo != "golang/go" {
		t.Fatalf("expected repo golang/go, got %s", svc.subscribeRepo)
	}

	body := decodeMap(t, w)
	if body["message"] != "Subscription successful. Confirmation email sent." {
		t.Fatalf("unexpected message: %q", body["message"])
	}
}

func TestSubscribe_InvalidRequestBody(t *testing.T) {
	svc := &fakeSubscriptionService{}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	if svc.subscribeCalled {
		t.Fatal("did not expect Subscribe to be called")
	}

	body := decodeMap(t, w)
	if body["error"] != "invalid request body" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_InvalidEmail(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: service.ErrInvalidEmail,
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"bad-email","repo":"golang/go"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrInvalidEmail.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_InvalidRepoFormat(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: service.ErrInvalidRepoFormat,
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"bad-format"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrInvalidRepoFormat.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_RepoNotFound(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: service.ErrRepositoryNotFound,
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"unknown/repo"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrRepositoryNotFound.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_AlreadySubscribed(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: service.ErrAlreadySubscribed,
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"golang/go"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrAlreadySubscribed.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_GitHubRateLimited(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: service.ErrGitHubRateLimited,
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"golang/go"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != "github api rate limit exceeded, please try again later" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestSubscribe_InternalError(t *testing.T) {
	svc := &fakeSubscriptionService{
		subscribeErr: errors.New("db down"),
	}
	router := setupTestRouter(svc)

	req := newJSONRequest(http.MethodPost, "/api/subscribe", `{"email":"test@example.com","repo":"golang/go"}`)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != "request cannot be processed" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestConfirm_Success(t *testing.T) {
	svc := &fakeSubscriptionService{}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/abc123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !svc.confirmCalled {
		t.Fatal("expected Confirm to be called")
	}

	if svc.confirmToken != "abc123" {
		t.Fatalf("expected token abc123, got %s", svc.confirmToken)
	}

	body := decodeMap(t, w)
	if body["message"] != "Subscription confirmed successfully" {
		t.Fatalf("unexpected message: %q", body["message"])
	}
}

func TestConfirm_InvalidToken(t *testing.T) {
	svc := &fakeSubscriptionService{
		confirmErr: service.ErrInvalidToken,
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/invalid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrInvalidToken.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestConfirm_TokenNotFound(t *testing.T) {
	svc := &fakeSubscriptionService{
		confirmErr: service.ErrTokenNotFound,
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/missing", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrTokenNotFound.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestConfirm_UnexpectedServiceError(t *testing.T) {
	svc := &fakeSubscriptionService{
		confirmErr: errors.New("unexpected"),
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/confirm/abc123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != "request cannot be processed" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestUnsubscribe_Success(t *testing.T) {
	svc := &fakeSubscriptionService{}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/unsub-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !svc.unsubscribeCalled {
		t.Fatal("expected Unsubscribe to be called")
	}

	if svc.unsubscribeToken != "unsub-token" {
		t.Fatalf("expected token unsub-token, got %s", svc.unsubscribeToken)
	}

	body := decodeMap(t, w)
	if body["message"] != "Unsubscribed successfully" {
		t.Fatalf("unexpected message: %q", body["message"])
	}
}

func TestUnsubscribe_InvalidToken(t *testing.T) {
	svc := &fakeSubscriptionService{
		unsubscribeErr: service.ErrInvalidToken,
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/invalid-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrInvalidToken.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestUnsubscribe_TokenNotFound(t *testing.T) {
	svc := &fakeSubscriptionService{
		unsubscribeErr: service.ErrTokenNotFound,
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/missing", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrTokenNotFound.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestUnsubscribe_InternalError(t *testing.T) {
	svc := &fakeSubscriptionService{
		unsubscribeErr: errors.New("db error"),
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/unsubscribe/missing", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != "request cannot be processed" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestGetSubscriptions_Success(t *testing.T) {
	svc := &fakeSubscriptionService{
		getSubscriptionsResult: []model.SubscriptionResponse{
			{Repo: "golang/go"},
			{Repo: "gin-gonic/gin"},
		},
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=test@example.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	if !svc.getSubscriptionsCalled {
		t.Fatal("expected GetSubscriptions to be called")
	}

	if svc.getSubscriptionsEmail != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", svc.getSubscriptionsEmail)
	}

	var body []model.SubscriptionResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if len(body) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(body))
	}
}

func TestGetSubscriptions_InvalidEmail(t *testing.T) {
	svc := &fakeSubscriptionService{
		getSubscriptionsErr: service.ErrInvalidEmail,
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=bad-email", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != service.ErrInvalidEmail.Error() {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}

func TestGetSubscriptions_InternalError(t *testing.T) {
	svc := &fakeSubscriptionService{
		getSubscriptionsErr: errors.New("db error"),
	}
	router := setupTestRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/subscriptions?email=test@example.com", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	body := decodeMap(t, w)
	if body["error"] != "request cannot be processed" {
		t.Fatalf("unexpected error message: %q", body["error"])
	}
}
