package auth

import (
	"context"
	"errors"
	"strconv"
)

type ctxKey int

// KeyValues is used to pass the SecurityPrincipal to the request context.Context
const KeyValues ctxKey = 1

// GetSecurityPrincipalFromContext returns the SecurityPrincipal from the request context.Context
func GetSecurityPrincipalFromContext(ctx context.Context) SecurityPrincipal {
	return ctx.Value(KeyValues).(SecurityPrincipal)
}

type Access uint32

const (
	AccessRead Access = 1 << (32 - 1 - iota)
	AccessCreate
	AccessUpdate
	AccessDelete
	AccessAdd
	AccessRemove
	AccessDisable
	AccessEnable
	AccessApprove

	AccessAll = AccessRead & AccessCreate & AccessUpdate & AccessDelete & AccessAdd & AccessRemove & AccessDisable & AccessEnable & AccessApprove
)

var ScopeSeparator = "/"

// A SecurityPrincipal represents any managed identity that is requesting access to a resource (a user, a service principal, etc)
type SecurityPrincipal interface {
	// Identity returns the principal identity
	Identity() string
	// HasRole returns true if the current SecurityPrincipal has assigned the role
	HasRole(roleName string) bool
	// HasPermission returns true if the current SecurityPrincipal has the Permission
	HasPermission(permission Permission) bool
}

// A Permission represents a set of resources and allowed Access
type Permission struct {
	Access Access
	// A Scope describes where an action can be performed
	// A scope might have many levels, and each level should be separated by a separator defined by ScopeSeparator
	// Scopes should be structured in a parent-child relationship. Each level of hierarchy makes the scope more specific
	// Examples:
	//	1. timesheet/team/team1 -> Allow the Access only to the timesheet of the team members from team1
	//	2. timesheet/*  -> Allow the Access to any timesheet from the organization
	//	3. *  -> Allow the Access to everything
	Scope string
}

// Implies returns true if the current Permission implies another Permission
func (p Permission) Implies(permission Permission) bool {
	//pScopeParts := strings.Split(p.Scope, ScopeSeparator)
	//ipScopeParts := strings.Split(permission.Scope, ScopeSeparator)

	//todo
	return false
}

func (p Permission) String() string {
	return p.Scope + ":" + strconv.Itoa(int(p.Access))
}

// A Role is a collection of allowed and denied permissions
// The denied permissions check has higher priority than allowed one
type Role struct {
	Name               string
	Description        string
	AllowedPermissions []Permission
	DeniedPermissions  []Permission
}

// HasPermission returns true if the current Role has the Permission
func (r Role) HasPermission(permission Permission) bool {
	for i := 0; i < len(r.DeniedPermissions); i++ {
		if r.DeniedPermissions[i].Implies(permission) {
			return false
		}
	}
	for i := 0; i < len(r.AllowedPermissions); i++ {
		if r.AllowedPermissions[i].Implies(permission) {
			return true
		}
	}
	return false
}

// A Group is a collection of roles
type Group struct {
	Name  string
	Roles []Role
}

// HasRole returns true if the current Group has the Role
func (g Group) HasRole(roleName string) bool {
	for i := 0; i < len(g.Roles); i++ {
		if roleName == g.Roles[i].Name {
			return true
		}
	}
	return false
}

// HasPermission returns true if the current Group has the Permission
func (g Group) HasPermission(permission Permission) bool {
	for i := 0; i < len(g.Roles); i++ {
		if g.Roles[i].HasPermission(permission) {
			return true
		}
	}
	return false
}

// A User implements SecurityPrincipal and represents an authenticated person
type User struct {
	Id               string
	Name             string
	IdentityPlatform string
	Groups           []Group
}

func (u User) Identity() string {
	return u.Id
}

func (u User) HasRole(roleName string) bool {
	for i := 0; i < len(u.Groups); i++ {
		if u.Groups[i].HasRole(roleName) {
			return true
		}
	}
	return false
}

func (u User) HasPermission(permission Permission) bool {
	for i := 0; i < len(u.Groups); i++ {
		if u.Groups[i].HasPermission(permission) {
			return true
		}
	}
	return false
}

// ParsePermission parse a string into a Permission
func ParsePermission(permissionAsString string) (Permission, error) {
	index := -1
	for i := len(permissionAsString) - 1; i >= 0; i-- {
		if i == ':' {
			index = i
			break
		}
	}
	if index == -1 {
		return Permission{}, errors.New("permission parsing failed, access field not found")
	}
	scope := permissionAsString[0:index]
	accessAsString := permissionAsString[index:]
	access, err := strconv.Atoi(accessAsString)
	if err != nil {
		return Permission{}, errors.New("permission parsing failed, access field is not a number")
	}
	return Permission{
		Access: Access(access),
		Scope:  scope,
	}, nil
}
