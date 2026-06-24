package handlers

import "github.com/adesubomi/pigeon-server/internal/app/services"

type Handler struct {
	authSvc     *services.AuthService
	deviceSvc   *services.DeviceService
	endpointSvc *services.EndpointService
	eventSvc    *services.EventService
	pushSvc     *services.PushService
}

func NewHandler(authSvc *services.AuthService) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

func (h *Handler) Auth() *AuthHandler {
	return new(AuthHandler(*h))
}

func (h *Handler) Endpoints() *EndpointHandler {
	return new(EndpointHandler(*h))
}

func (h *Handler) Devices() *DeviceHandler {
	return new(DeviceHandler(*h))
}

func (h *Handler) Events() *EventHandler {
	return new(EventHandler(*h))
}
