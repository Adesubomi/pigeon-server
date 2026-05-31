package push

import (
	"sync"
	"time"

	"github.com/adesubomi/pigeon-server/internal/domain/device"
)

type Message struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type DeliveryAttempt struct {
	DeviceID    string
	Status      string
	Error       string
	DeliveredAt *time.Time
}

type Hub struct {
	mu      sync.RWMutex
	streams map[string]map[string]chan Message
}

func NewHub() *Hub {
	return &Hub{streams: make(map[string]map[string]chan Message)}
}

func (h *Hub) Register(device *device.Device) (<-chan Message, func()) {
	ch := make(chan Message, 16)

	h.mu.Lock()
	if h.streams[device.EndpointID] == nil {
		h.streams[device.EndpointID] = make(map[string]chan Message)
	}
	h.streams[device.EndpointID][device.ID] = ch
	h.mu.Unlock()

	unregister := func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if endpointStreams, ok := h.streams[device.EndpointID]; ok {
			delete(endpointStreams, device.ID)
			if len(endpointStreams) == 0 {
				delete(h.streams, device.EndpointID)
			}
		}
		close(ch)
	}

	return ch, unregister
}

func (h *Hub) SendToEndpoint(endpointID string, message Message) []DeliveryAttempt {
	h.mu.RLock()
	endpointStreams := h.streams[endpointID]
	attempts := make([]DeliveryAttempt, 0, len(endpointStreams))
	for deviceID, ch := range endpointStreams {
		now := time.Now().UTC()
		select {
		case ch <- message:
			attempts = append(attempts, DeliveryAttempt{
				DeviceID:    deviceID,
				Status:      "delivered",
				DeliveredAt: &now,
			})
		default:
			attempts = append(attempts, DeliveryAttempt{
				DeviceID: deviceID,
				Status:   "failed",
				Error:    "device stream buffer is full",
			})
		}
	}
	h.mu.RUnlock()

	return attempts
}
