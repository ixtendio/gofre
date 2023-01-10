package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

type ctxKey int

// SecurityPrincipalCtxKey is used to pass the SecurityPrincipal to the request context.Context
const SecurityPrincipalCtxKey ctxKey = 1

// GetSecurityPrincipalFromContext returns the SecurityPrincipal from the request context.Context
func GetSecurityPrincipalFromContext(ctx context.Context) SecurityPrincipal {
	if sp, ok := ctx.Value(SecurityPrincipalCtxKey).(SecurityPrincipal); ok {
		return sp
	}
	return nil
}

type Access uint32

const (
	AccessRead Access = 1 << iota
	AccessCreate
	AccessUpdate
	AccessDelete
	AccessAdd
	AccessRemove
	AccessDisable
	AccessEnable
	AccessApprove
	accessUnknown
)

func AccessAll() Access {
	var all Access
	for i := 1 << 0; Access(i) < accessUnknown; i = i << 1 {
		all |= Access(i)
	}
	return all
}

var ScopeSeparator = '/'

// A SecurityPrincipal represents any managed identity that is requesting access to a resource (a user, a service principal, etc)
type SecurityPrincipal interface {
	// Identity returns the principal identity
	Identity() string
	// HasRole returns true if the current SecurityPrincipal has assigned the role
	HasRole(roleName string) bool
	// HasPermission returns true if the current SecurityPrincipal has the Permission
	HasPermission(permission Permission) bool
	// String returns a string representation of the SecurityPrincipal
	String() string
}

// A Permission has a Scope and Access. A Scope describes where an action can be performed
// For simplicity, the scope might have maximum 3 levels, (domain, subdomain and resource) separated by ScopeSeparator
// Scopes should be structured in a parent-child relationship. Each level of hierarchy makes the scope more specific
//
// Examples:
//  1. admin/timesheet/team1 -> Allow access only to the resource team1 from admin/timesheet
//  2. admin/timesheet/*     -> Allow access to all resources from admin/timesheet
//  3. admin/*/team1 		 -> Allow access to all subdomains from the admin domain related to the resource team1
//  4. admin/*   			 -> Allow access to all subdomains and all resources from the admin domain
//  5. *  					 -> Allow access to all domains
type Permission struct {
	Scope  string
	Access Access
}

func AllPermissions() Permission {
	return Permission{
		Scope:  "*",
		Access: AccessAll(),
	}
}

func NewPermission(scope string, access Access) (Permission, error) {
	var separatorsCount int
	for _, c := range scope {
		if c == ScopeSeparator {
			separatorsCount++
		}
		if separatorsCount > 2 {
			return Permission{}, errors.New("the scope should have maximum 3 levels")
		}
	}
	return Permission{scope, access}, nil
}

// Implies returns true if the current Permission implies anotherPermission
// This function assumes that the scope of the Permission from the argument, does not contain the wildcard (*)
func (p Permission) Implies(anotherPermission Permission) bool {
	sep := string(ScopeSeparator)
	pScopeParts := strings.Split(p.Scope, sep)
	ipScopeParts := strings.Split(anotherPermission.Scope, sep)
	if len(ipScopeParts) < len(pScopeParts) {
		return false
	}

	if p.Access&anotherPermission.Access != anotherPermission.Access {
		return false
	}

	var s string
	var j int
	for i := 0; i < len(ipScopeParts); i++ {
		si := ipScopeParts[i]
		if j < len(pScopeParts) {
			s = pScopeParts[j]
			j++
		}

		if s != "*" && s != si {
			return false
		}
	}
	return true
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
	// the user internal id
	Id string
	// the name of user
	Name string
	// the id/name of the platform were the user was authenticated (for example Google, Linkedin, Internal, etc)
	IdentityPlatform string
	// the security groups where this user belongs
	Groups []Group
	// a field where any additional data to this user can be attached
	Attachment any
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

func (u User) String() string {
	return u.Name
}

// ParsePermission parse a string into a Permission
func ParsePermission(permissionAsString string) (Permission, error) {
	index := -1
	for i := len(permissionAsString) - 1; i >= 0; i-- {
		if permissionAsString[i] == ':' {
			index = i
			break
		}
	}
	if index == -1 || index == len(permissionAsString)-1 {
		return Permission{}, errors.New("permission parsing failed, access field not found")
	}
	scope := permissionAsString[0:index]
	accessAsString := permissionAsString[index+1:]
	access, err := strconv.ParseUint(accessAsString, 10, 32)
	if err != nil {
		return Permission{}, errors.New("permission parsing failed, " + err.Error())
	}

	return NewPermission(scope, Access(access))
}
