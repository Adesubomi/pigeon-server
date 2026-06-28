package auth

type SessionMeta struct {
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	DeviceID  string `json:"device_id,omitempty"`
}
