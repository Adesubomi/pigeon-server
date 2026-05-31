package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adesubomi/pigeon-server/config"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"gorm.io/gorm"
)

type contextKey string

const userContextKey contextKey = "auth.user"

type Service struct {
	db         *gorm.DB
	cfg        *config.Config
	httpClient *http.Client
}

func NewService(db *gorm.DB, cfg *config.Config) *Service {
	return &Service{
		db:         db,
		cfg:        cfg,
		httpClient: http.DefaultClient,
	}
}

func (s *Service) GitHubLoginURL(ctx context.Context, redirectURI, state string) (string, error) {
	if s.cfg.GitHubClientID == "" {
		return "", apperr.NotImplemented()
	}

	values := url.Values{}
	values.Set("client_id", s.cfg.GitHubClientID)
	values.Set("scope", "user:email")
	if redirectURI != "" {
		values.Set("redirect_uri", redirectURI)
	}
	if state != "" {
		values.Set("state", state)
	}

	return "https://github.com/login/oauth/authorize?" + values.Encode(), nil
}

func (s *Service) ExchangeGitHubCode(ctx context.Context, input GitHubExchangeInput) (*AuthTokenResponse, error) {
	if strings.TrimSpace(input.Code) == "" {
		return nil, apperr.BadRequest("auth.code_required", "OAuth code is required")
	}
	if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
		return nil, apperr.NotImplemented()
	}

	githubToken, err := s.exchangeGitHubToken(ctx, input)
	if err != nil {
		return nil, err
	}

	githubUser, err := s.fetchGitHubUser(ctx, githubToken.AccessToken)
	if err != nil {
		return nil, err
	}

	user := User{
		GitHubID:  fmt.Sprintf("%d", githubUser.ID),
		Email:     githubUser.Email,
		Name:      githubUser.Name,
		AvatarURL: githubUser.AvatarURL,
	}
	if user.Name == "" {
		user.Name = githubUser.Login
	}

	if err := s.db.WithContext(ctx).
		Where(User{GitHubID: user.GitHubID}).
		Assign(User{
			Email:     user.Email,
			Name:      user.Name,
			AvatarURL: user.AvatarURL,
		}).
		FirstOrCreate(&user).Error; err != nil {
		return nil, apperr.Internal(err)
	}

	return s.tokenResponse(&user)
}

func (s *Service) CurrentUser(ctx context.Context) (*User, error) {
	user, ok := UserFromContext(ctx)
	if !ok {
		return nil, apperr.Unauthorized("auth.unauthorized", "Authentication required")
	}
	return user, nil
}

func (s *Service) RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawToken := bearerToken(r)
		if rawToken == "" {
			respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
			return
		}

		userID, ok := s.verifyAccessToken(rawToken)
		if !ok {
			respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
			return
		}

		var user User
		if err := s.db.WithContext(r.Context()).First(&user, "id = ?", userID).Error; err != nil {
			respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
			return
		}

		next.ServeHTTP(w, r.WithContext(ContextWithUser(r.Context(), &user)))
	})
}

func (s *Service) CreateAccessToken(userID string) (string, time.Time, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(s.cfg.AuthAccessTokenTTL)

	header, err := encodeJWTPart(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", time.Time{}, err
	}
	claims, err := encodeJWTPart(accessTokenClaims{
		Subject:   userID,
		TokenType: "user",
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	})
	if err != nil {
		return "", time.Time{}, err
	}

	signingInput := header + "." + claims
	signature := s.sign(signingInput)
	return signingInput + "." + signature, expiresAt, nil
}

func (s *Service) CreateSession(userID string) string {
	token, _, err := s.CreateAccessToken(userID)
	if err != nil {
		return ""
	}
	return token
}

func (s *Service) tokenResponse(user *User) (*AuthTokenResponse, error) {
	accessToken, expiresAt, err := s.CreateAccessToken(user.ID)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &AuthTokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(expiresAt).Seconds()),
		User:        user,
	}, nil
}

func (s *Service) verifyAccessToken(rawToken string) (string, bool) {
	parts := strings.Split(rawToken, ".")
	if len(parts) != 3 {
		return "", false
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := s.sign(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return "", false
	}

	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if !decodeJWTPart(parts[0], &header) || header.Algorithm != "HS256" || header.Type != "JWT" {
		return "", false
	}

	var claims accessTokenClaims
	if !decodeJWTPart(parts[1], &claims) {
		return "", false
	}
	if claims.Subject == "" || claims.TokenType != "user" {
		return "", false
	}
	if time.Now().UTC().Unix() >= claims.ExpiresAt {
		return "", false
	}

	return claims.Subject, true
}

func ContextWithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

func (s *Service) exchangeGitHubToken(ctx context.Context, input GitHubExchangeInput) (*githubTokenResponse, error) {
	payload := url.Values{}
	payload.Set("client_id", s.cfg.GitHubClientID)
	payload.Set("client_secret", s.cfg.GitHubClientSecret)
	payload.Set("code", strings.TrimSpace(input.Code))
	if input.RedirectURI != "" {
		payload.Set("redirect_uri", input.RedirectURI)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://github.com/login/oauth/access_token", strings.NewReader(payload.Encode()))
	if err != nil {
		return nil, apperr.Internal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, apperr.Unauthorized("auth.github_exchange_failed", "GitHub OAuth exchange failed")
	}

	var tokenResponse githubTokenResponse
	if err := json.Unmarshal(resBody, &tokenResponse); err != nil {
		return nil, apperr.Internal(err)
	}
	if tokenResponse.Error != "" {
		return nil, apperr.Unauthorized("auth.github_exchange_failed", "GitHub OAuth exchange failed")
	}
	if tokenResponse.AccessToken == "" {
		return nil, apperr.Unauthorized("auth.github_token_missing", "GitHub OAuth token missing")
	}

	return &tokenResponse, nil
}

func (s *Service) fetchGitHubUser(ctx context.Context, accessToken string) (*githubUserResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user", nil)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, apperr.Unauthorized("auth.github_user_failed", "GitHub user lookup failed")
	}

	var user githubUserResponse
	if err := json.Unmarshal(resBody, &user); err != nil {
		return nil, apperr.Internal(err)
	}
	if user.ID == 0 {
		return nil, apperr.Unauthorized("auth.github_user_invalid", "GitHub user lookup failed")
	}

	return &user, nil
}

func (s *Service) sign(value string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.AppKey))
	_, _ = mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeJWTPart(value any) (string, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeJWTPart(value string, target any) bool {
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return false
	}
	return json.Unmarshal(payload, target) == nil
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	prefix, value, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(prefix, "Bearer") {
		return ""
	}
	return strings.TrimSpace(value)
}

type accessTokenClaims struct {
	Subject   string `json:"sub"`
	TokenType string `json:"typ"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type githubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

type githubUserResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}
