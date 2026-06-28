package auth

import (
	"fmt"
)

type PermSubject string
type PermAction string

const DefaultRoleSuper = "Owner"

const (
	PermSubjEndpoints PermSubject = "endpoints"
)

const (
	PermActionCreate PermAction = "create"
	PermActionView   PermAction = "view"
)

type Permission struct {
	Subject PermSubject `json:"subject"`
	Action  PermAction  `json:"action"`
}

func (p Permission) IsValid() bool {
	reg, ok := AvailablePermissions[p.Subject]
	if !ok {
		return false
	}
	_, ok = reg.Actions[p.Action]
	return ok
}

func (p Permission) Code() string {
	return PermissionCode(p.Subject, p.Action)
}

type PermissionRegistration struct {
	IsActive    *bool  `json:"isActive,omitempty"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

type PermissionSet map[PermSubject]PermissionSubjectActionRegistration

type PermissionSubjectActionRegistration struct {
	Actions map[PermAction]PermissionRegistration `json:"actions"`
}

var AvailablePermissions = PermissionSet{
	PermSubjEndpoints: PermissionSubjectActionRegistration{
		Actions: map[PermAction]PermissionRegistration{
			PermActionView:   {Description: "Can view endpoints"},
			PermActionCreate: {Description: "Can create endpoints"},
		},
	},
}

func PermissionCode(subject PermSubject, action PermAction) string {
	return fmt.Sprintf("%s.%s", subject, action)
}
