package auth

import (
	"encoding/json"
	"net/http"

	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url, err := h.service.GitHubLoginURL(
		r.Context(),
		r.URL.Query().Get("redirect_uri"),
		r.URL.Query().Get("state"),
	)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, GitHubLoginResponse{URL: url})
}

func (h *Handler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ExchangeGitHubCode(r.Context(), GitHubExchangeInput{
		Code:        r.URL.Query().Get("code"),
		RedirectURI: r.URL.Query().Get("redirect_uri"),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) GitHubExchange(w http.ResponseWriter, r *http.Request) {
	var req GitHubExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.service.ExchangeGitHubCode(r.Context(), GitHubExchangeInput{
		Code:        req.Code,
		RedirectURI: req.RedirectURI,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	respond.NoContent(w)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.service.CurrentUser(r.Context())
	if err != nil {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	respond.OK(w, MeResponse{User: user})
}
