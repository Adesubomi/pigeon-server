package endpoint

import "time"

type CreateEndpointRequest struct {
	Name string `json:"name"`
}

type UpdateEndpointRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}

type CreateEndpointInput struct {
	UserID string
	Name   string
}

type UpdateEndpointInput struct {
	UserID   string
	ID       string
	Name     *string
	IsActive *bool
}

type EndpointResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PairingCodeResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type DeviceSummary struct {
	ID         string     `json:"id"`
	DeviceID   string     `json:"device_id"`
	DeviceName string     `json:"device_name"`
	IsActive   bool       `json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
}

type EventSummary struct {
	ID          string    `json:"id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	ContentType string    `json:"content_type"`
	ReceivedAt  time.Time `json:"received_at"`
	CreatedAt   time.Time `json:"created_at"`
}
