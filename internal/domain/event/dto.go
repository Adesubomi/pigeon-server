package event

import (
	"net/http"
	"net/url"
	"time"
)

type ReceiveWebhookInput struct {
	Slug        string
	Method      string
	Path        string
	Headers     http.Header
	Query       url.Values
	Body        []byte
	ContentType string
}

type WebhookReceivedResponse struct {
	EventID string `json:"event_id"`
}

type EventResponse struct {
	ID          string    `json:"id"`
	EndpointID  string    `json:"endpoint_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Headers     any       `json:"headers"`
	Query       any       `json:"query"`
	Body        string    `json:"body"`
	ContentType string    `json:"content_type"`
	ReceivedAt  time.Time `json:"received_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type ReplayEventInput struct {
	UserID  string
	EventID string
}

type ReplayEventResponse struct {
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

type PushPayload struct {
	EventID     string    `json:"event_id"`
	EndpointID  string    `json:"endpoint_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Headers     any       `json:"headers"`
	Query       any       `json:"query"`
	Body        string    `json:"body"`
	ContentType string    `json:"content_type"`
	ReceivedAt  time.Time `json:"received_at"`
}
