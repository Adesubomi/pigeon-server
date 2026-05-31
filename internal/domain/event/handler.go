package event

import (
	"io"
	"net/http"

	"github.com/adesubomi/pigeon-server/internal/domain/auth"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ReceiveWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10<<20))
	if err != nil {
		respond.Error(w, apperr.BadRequest("request.body_too_large", "Request body is too large"))
		return
	}

	result, err := h.service.ReceiveWebhook(r.Context(), ReceiveWebhookInput{
		Slug:        chi.URLParam(r, "slug"),
		Method:      r.Method,
		Path:        r.URL.Path,
		Headers:     r.Header.Clone(),
		Query:       r.URL.Query(),
		Body:        body,
		ContentType: r.Header.Get("Content-Type"),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.GetEvent(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) ReplayEvent(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.ReplayEvent(r.Context(), ReplayEventInput{
		UserID:  user.ID,
		EventID: chi.URLParam(r, "id"),
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}
