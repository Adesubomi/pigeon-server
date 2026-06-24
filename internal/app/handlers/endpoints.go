package handlers

import (
	"encoding/json"
	"net/http"

	endpointDomain "github.com/adesubomi/pigeon-server/internal/domain/endpoint"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
	"github.com/go-chi/chi/v5"
)

type EndpointHandler Handler

func (h *EndpointHandler) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}

	var req endpointDomain.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.endpointSvc.CreateEndpoint(r.Context(), endpointDomain.CreateEndpointInput{UserID: user.ID, Name: req.Name})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *EndpointHandler) ListEndpoints(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.endpointSvc.ListEndpoints(r.Context(), user.ID)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *EndpointHandler) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.endpointSvc.GetEndpoint(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *EndpointHandler) UpdateEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}

	var req endpointDomain.UpdateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.endpointSvc.UpdateEndpoint(r.Context(), endpointDomain.UpdateEndpointInput{
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

func (h *EndpointHandler) DeleteEndpoint(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	if err := h.endpointSvc.DeleteEndpoint(r.Context(), user.ID, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, err)
		return
	}
	respond.NoContent(w)
}

func (h *EndpointHandler) GeneratePairingCode(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.endpointSvc.GeneratePairingCode(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *EndpointHandler) ListEndpointDevices(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.endpointSvc.ListEndpointDevices(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *EndpointHandler) ListEndpointEvents(w http.ResponseWriter, r *http.Request) {
	user, ok := h.authSvc.UserFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("auth.unauthorized", "Authentication required"))
		return
	}
	result, err := h.endpointSvc.ListEndpointEvents(r.Context(), user.ID, chi.URLParam(r, "id"))
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}
