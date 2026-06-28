package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/adesubomi/pigeon-server/config"
	authDomain "github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
)

func TestGitHubLoginURLIncludesValidatedRedirectAndState(t *testing.T) {
	service := newTestService()

	loginURL, err := service.GitHubLoginURL(context.Background(), "http://localhost:3000/auth/callback", "random-state")
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := url.Parse(loginURL)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	if got := query.Get("client_id"); got != "client-id" {
		t.Fatalf("client_id = %q", got)
	}
	if got := query.Get("redirect_uri"); got != "http://localhost:3000/auth/callback" {
		t.Fatalf("redirect_uri = %q", got)
	}
	if got := query.Get("state"); got != "random-state" {
		t.Fatalf("state = %q", got)
	}
}

func TestGitHubLoginURLRejectsUntrustedRedirect(t *testing.T) {
	service := newTestService()

	_, err := service.GitHubLoginURL(context.Background(), "https://attacker.example/callback", "state")
	assertAppErrorCode(t, err, "auth.redirect_uri_invalid")
}

func TestGitHubLoginURLUsesConfiguredDefaultRedirect(t *testing.T) {
	service := newTestService()

	loginURL, err := service.GitHubLoginURL(context.Background(), "", "state")
	if err != nil {
		t.Fatal(err)
	}
	parsed, _ := url.Parse(loginURL)
	if got := parsed.Query().Get("redirect_uri"); got != "http://localhost:3000/auth/callback" {
		t.Fatalf("redirect_uri = %q", got)
	}
}

func TestGitHubLoginURLRequiresConfiguration(t *testing.T) {
	service := newTestService()
	service.cfg.GitHubClientID = ""

	_, err := service.GitHubLoginURL(context.Background(), "", "")
	assertAppErrorCode(t, err, "auth.github_not_configured")
}

func TestExchangeGitHubTokenSendsValidatedConfiguration(t *testing.T) {
	service := newTestService()
	service.httpClient = &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatal(err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Fatal(err)
		}
		if got := values.Get("client_secret"); got != "client-secret" {
			t.Fatalf("client_secret = %q", got)
		}
		if got := values.Get("code"); got != "oauth-code" {
			t.Fatalf("code = %q", got)
		}
		if got := values.Get("redirect_uri"); got != "http://localhost:3000/auth/callback" {
			t.Fatalf("redirect_uri = %q", got)
		}
		return jsonResponse(http.StatusOK, `{"access_token":"github-token","token_type":"bearer"}`), nil
	})}

	token, err := service.exchangeGitHubToken(context.Background(), authDomain.GitHubExchangeInput{
		Code:        "oauth-code",
		RedirectURI: "http://localhost:3000/auth/callback",
	})
	if err != nil {
		t.Fatal(err)
	}
	if token.AccessToken != "github-token" {
		t.Fatalf("access token = %q", token.AccessToken)
	}
}

func TestExchangeGitHubCodeValidatesInputAndConfiguration(t *testing.T) {
	service := newTestService()

	_, err := service.ExchangeGitHubCode(context.Background(), authDomain.GitHubExchangeInput{})
	assertAppErrorCode(t, err, "auth.code_required")

	service.cfg.GitHubClientSecret = ""
	_, err = service.ExchangeGitHubCode(context.Background(), authDomain.GitHubExchangeInput{Code: "code"})
	assertAppErrorCode(t, err, "auth.github_not_configured")

	service.cfg.GitHubClientSecret = "client-secret"
	_, err = service.ExchangeGitHubCode(context.Background(), authDomain.GitHubExchangeInput{
		Code:        "code",
		RedirectURI: "https://attacker.example/callback",
	})
	assertAppErrorCode(t, err, "auth.redirect_uri_invalid")
}

func TestAccessTokenAndRequireUser(t *testing.T) {
	service := newTestService()
	expectedUser := &authDomain.User{ID: "user-123", Name: "Ada", Email: "ada@example.com"}
	service.loadUser = func(_ context.Context, userID string) (*authDomain.User, error) {
		if userID != expectedUser.ID {
			t.Fatalf("user ID = %q", userID)
		}
		return expectedUser, nil
	}

	token, expiresAt, err := service.CreateAccessToken(expectedUser.ID)
	if err != nil {
		t.Fatal(err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Fatal("token expiry is not in the future")
	}
	if subject, ok := service.verifyAccessToken(token); !ok || subject != expectedUser.ID {
		t.Fatalf("verified subject = %q, ok = %v", subject, ok)
	}

	handler := service.RequireUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := UserFromContext(r.Context())
		if !ok || user.ID != expectedUser.ID {
			t.Fatal("authenticated user missing from context")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", res.Code, res.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer invalid")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token status = %d", res.Code)
	}
}

func TestCurrentUserReturnsAuthenticatedUser(t *testing.T) {
	service := newTestService()
	user := &authDomain.User{ID: "user-123", Name: "Ada"}
	ctx := ContextWithUser(context.Background(), user)

	result, err := service.CurrentUser(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != user.ID {
		t.Fatalf("user ID = %q", result.ID)
	}
}

func newTestService() *AuthService {
	return NewAuth(nil, &config.Config{
		AppKey:             "test-signing-key",
		GitHubClientID:     "client-id",
		GitHubClientSecret: "client-secret",
		WebAppURL:          "http://localhost:3000",
		AuthAccessTokenTTL: time.Hour,
	})
}

func assertAppErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	appErr, ok := err.(*apperr.AppError)
	if !ok || appErr.Code != code {
		t.Fatalf("error = %#v, want code %q", err, code)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
