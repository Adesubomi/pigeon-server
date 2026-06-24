package handlers

import (
	"encoding/json"
	"net/http"

	domain "github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
)

type AuthHandler Handler

func (h *AuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url, err := h.authSvc.GitHubLoginURL(
		r.Context(),
		r.URL.Query().Get("redirect_uri"),
		r.URL.Query().Get("state"),
	)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, domain.GitHubLoginResponse{URL: url})
}

func (h *AuthHandler) GitHubExchange(w http.ResponseWriter, r *http.Request) {
	var req domain.GitHubExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.authSvc.ExchangeGitHubCode(r.Context(), domain.GitHubExchangeInput{
		Code:        req.Code,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	respond.NoContent(w)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.authSvc.CurrentUser(r.Context())
	if err != nil {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	respond.OK(w, domain.MeResponse{User: user})
}
