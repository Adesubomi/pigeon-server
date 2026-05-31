package endpoint

import (
	"encoding/json"
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

func (h *Handler) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}

	var req CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.service.CreateEndpoint(r.Context(), CreateEndpointInput{UserID: user.ID, Name: req.Name})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *Handler) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.ListEndpoints(r.Context(), user.ID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.GetEndpoint(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}

	var req UpdateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.service.UpdateEndpoint(r.Context(), UpdateEndpointInput{
		UserID:   user.ID,
		ID:       chi.URLParam(r, "id"),
		Name:     req.Name,
		IsActive: req.IsActive,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	if err := h.service.DeleteEndpoint(r.Context(), user.ID, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, err)
		return
	}
	respond.NoContent(w)
}

func (h *Handler) GeneratePairingCode(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.GeneratePairingCode(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *Handler) ListEndpointDevices(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.ListEndpointDevices(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) ListEndpointEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.service.ListEndpointEvents(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}
