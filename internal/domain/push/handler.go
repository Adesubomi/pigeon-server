package push

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/adesubomi/pigeon-server/internal/domain/device"
	"github.com/adesubomi/pigeon-server/pkg/apperr"
	"github.com/adesubomi/pigeon-server/pkg/respond"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Stream(w http.ResponseWriter, r *http.Request) {
	currentDevice, ok := device.DeviceFromContext(r.Context())
	if !ok {
		respond.Error(w, apperr.Unauthorized("device.unauthorized", "Device authentication required"))
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		respond.Error(w, apperr.Internal(fmt.Errorf("streaming is not supported by response writer")))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unregister := h.service.Hub().Register(currentDevice)
	defer unregister()

	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			payload, err := json.Marshal(msg.Data)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", msg.Event, payload)
			flusher.Flush()
		}
	}
}
