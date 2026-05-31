package device

import "time"

type PairDeviceRequest struct {
	Code       string `json:"code"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type PairDeviceInput struct {
	Code       string
	DeviceID   string
	DeviceName string
}

type PairDeviceResponse struct {
	DeviceID string `json:"device_id"`
	Token    string `json:"token"`
}

type HeartbeatResponse struct {
	LastSeenAt time.Time `json:"last_seen_at"`
}

type UpdateDeviceRequest struct {
	DeviceName *string `json:"device_name"`
	IsActive   *bool   `json:"is_active"`
}

type UpdateDeviceInput struct {
	ID         string
	Device     *Device
	DeviceName *string
	IsActive   *bool
}

type DeviceResponse struct {
	ID         string     `json:"id"`
	EndpointID string     `json:"endpoint_id"`
	DeviceID   string     `json:"device_id"`
	DeviceName string     `json:"device_name"`
	IsActive   bool       `json:"is_active"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}
