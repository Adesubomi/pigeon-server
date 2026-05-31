package device

import (
	"encoding/json"
	"net/http"

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

func (h *Handler) PairDevice(w http.ResponseWriter, r *http.Request) {
	var req PairDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.service.PairDevice(r.Context(), PairDeviceInput{
		Code:       req.Code,
		DeviceID:   req.DeviceID,
		DeviceName: req.DeviceName,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.Created(w, result)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	currentDevice, ok := DeviceFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
		return
	}
	result, err := h.service.Heartbeat(r.Context(), currentDevice)
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) UpdateDevice(w http.ResponseWriter, r *http.Request) {
	currentDevice, ok := DeviceFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
		return
	}

	var req UpdateDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.Error(w, apperr.BadRequest("request.invalid_json", "Invalid JSON body"))
		return
	}

	result, err := h.service.UpdateDevice(r.Context(), UpdateDeviceInput{
		ID:         chi.URLParam(r, "id"),
		Device:     currentDevice,
		DeviceName: req.DeviceName,
		IsActive:   req.IsActive,
	})
	if err != nil {
		respond.Error(w, err)
		return
	}
	respond.OK(w, result)
}

func (h *Handler) DeleteDevice(w http.ResponseWriter, r *http.Request) {
	currentDevice, ok := DeviceFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
		return
	}
	if err := h.service.DeleteDevice(r.Context(), currentDevice, chi.URLParam(r, "id")); err != nil {
		respond.Error(w, err)
		return
	}
	respond.NoContent(w)
}
