package auth

import "time"

type ActorRole struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Tag         string       `json:"tag"`
	IsReadOnly  bool         `json:"isReadOnly"`
	Permissions []Permission `json:"permissions"`
}

type ACL struct {
	WorkspaceID   string `json:"workspaceId"`
	WorkspaceName string `json:"workspaceName"`

	Scope       Scope        `json:"scope,omitempty"`
	Roles       []ActorRole  `json:"roles,omitempty"`
	Permissions []Permission `json:"permissions,omitempty"`
}

func (a ACL) IsSuper() bool {
	for _, role := range a.Roles {
		if role.Tag == DefaultRoleSuper {
			return true
		}
	}
	return false
}

func (a ACL) Can(action PermAction, subject PermSubject) bool {
	if a.IsSuper() {
		return true
	}

	for _, perm := range a.Permissions {
		if perm.Subject == subject && perm.Action == action {
			return true
		}
	}
	return false
}

type AuthorizationSnapshot = ACL

type LiveSessionUser struct {
	ID            string `json:"id"`
	WorkspaceID   string `json:"workspaceId"`
	WorkspaceName string `json:"workspaceName"`
	ACL           ACL    `json:"acl"`
}

type LiveSession struct {
	ID         string          `json:"id"`
	User       LiveSessionUser `json:"userId"`
	Meta       SessionMeta     `json:"meta,omitempty"`
	IssuedAt   time.Time       `json:"issuedAt"`
	ExpiresAt  time.Time       `json:"expiresAt"`
	LastSeenAt time.Time       `json:"last_seenAt,omitempty"`
}
