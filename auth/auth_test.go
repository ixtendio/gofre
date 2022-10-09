package auth

import (
	"context"
	"math"
	"reflect"
	"strconv"
	"testing"
)

func TestPermission_Implies(t *testing.T) {
	type fields struct {
		Access Access
		Scope  string
	}
	tests := []struct {
		name   string
		fields fields
		args   []Permission
		want   bool
	}{
		{
			name: "fix matching with access => true",
			fields: fields{
				Access: AccessCreate | AccessRead | AccessUpdate,
				Scope:  "domain/subdomain/resource",
			},
			args: []Permission{
				{"domain/subdomain/resource", AccessCreate},
				{"domain/subdomain/resource", AccessRead},
				{"domain/subdomain/resource", AccessUpdate},
				{"domain/subdomain/resource", AccessCreate | AccessRead},
				{"domain/subdomain/resource", AccessCreate | AccessUpdate},
				{"domain/subdomain/resource", AccessRead | AccessUpdate},
				{"domain/subdomain/resource", AccessCreate | AccessRead | AccessUpdate},
			},
			want: true,
		},
		{
			name: "fix matching with access => false",
			fields: fields{
				Access: AccessCreate | AccessRead | AccessUpdate,
				Scope:  "domain/subdomain/resource",
			},
			args: []Permission{
				{"domain/subdomain/resource", AccessDelete},
				{"domain/subdomain/resource", AccessCreate | AccessDelete},
				{"domain/subdomain/resource", AccessRead | AccessApprove},
				{"domain/subdomain/resource", AccessAll()},
			},
			want: false,
		},
		{
			name: "fix match when arg levels count is less than permissions levels count => false",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/subdomain/resource",
			},
			args: []Permission{{"domain/subdomain", AccessAll()}},
			want: false,
		},
		{
			name: "match when subdomain is wildcard => false",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/*/resource",
			},
			args: []Permission{{"domain/subdomain/resource1", AccessAll()}},
			want: false,
		},
		{
			name: "fix matching with all access => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/subdomain/resource",
			},
			args: []Permission{
				{"domain/subdomain/resource", AccessCreate},
				{"domain/subdomain/resource", AccessDelete},
				{"domain/subdomain/resource", AccessCreate | AccessDelete},
				{"domain/subdomain/resource", AccessRead | AccessApprove},
				{"domain/subdomain/resource", AccessAll()},
			},
			want: true,
		},
		{
			name: "fix matching => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/subdomain/resource",
			},
			args: []Permission{{"domain/subdomain/resource", AccessAll()}},
			want: true,
		},
		{
			name: "match when resource is wildcard => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/subdomain/*",
			},
			args: []Permission{{"domain/subdomain/resource", AccessAll()}},
			want: true,
		},
		{
			name: "match when subdomain is wildcard => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/*/resource",
			},
			args: []Permission{{"domain/subdomain/resource", AccessAll()}},
			want: true,
		},
		{
			name: "match when after domain is wildcard => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "domain/*",
			},
			args: []Permission{{"domain/subdomain/resource", AccessAll()}},
			want: true,
		},
		{
			name: "match everything => true",
			fields: fields{
				Access: AccessAll(),
				Scope:  "*",
			},
			args: []Permission{{"domain/subdomain/resource", AccessAll()}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Permission{
				Access: tt.fields.Access,
				Scope:  tt.fields.Scope,
			}
			for _, arg := range tt.args {
				if got := p.Implies(arg); got != tt.want {
					t.Errorf("Implies() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNewPermission(t *testing.T) {
	type args struct {
		scope  string
		access Access
	}
	tests := []struct {
		name    string
		args    args
		want    Permission
		wantErr bool
	}{
		{
			name: "parse 4 levels => error",
			args: args{
				scope:  "a/b/c/d",
				access: 0,
			},
			want:    Permission{},
			wantErr: true,
		},
		{
			name: "parse 3 levels => ok",
			args: args{
				scope:  "a/b/c",
				access: 0,
			},
			want: Permission{
				Scope:  "a/b/c",
				Access: 0,
			},
			wantErr: false,
		},
		{
			name: "parse 2 levels => ok",
			args: args{
				scope:  "a/b",
				access: 0,
			},
			want: Permission{
				Scope:  "a/b",
				Access: 0,
			},
			wantErr: false,
		},
		{
			name: "parse 1 level => ok",
			args: args{
				scope:  "a",
				access: 0,
			},
			want: Permission{
				Scope:  "a",
				Access: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPermission(tt.args.scope, tt.args.access)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPermission() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePermission(t *testing.T) {
	tests := []struct {
		name    string
		args    string
		want    Permission
		wantErr bool
	}{
		{
			name:    "permission with correct access => ok",
			args:    "a/b/c:" + strconv.Itoa(int(AccessCreate|AccessUpdate)),
			want:    Permission{"a/b/c", AccessCreate | AccessUpdate},
			wantErr: false,
		},
		{
			name:    "permission without access (1) => error",
			args:    "a/b/c:",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "permission without access (2) => error",
			args:    "a/b/c",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "permission with string access => error",
			args:    "a/b/c:abc",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "permission with negative access => error",
			args:    "a/b/c:-1",
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "permission with int64 access => error",
			args:    "a/b/c:" + strconv.Itoa(math.MaxInt64),
			want:    Permission{},
			wantErr: true,
		},
		{
			name:    "permission with float access => error",
			args:    "a/b/c:1.23",
			want:    Permission{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePermission(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePermission() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermission_String(t *testing.T) {
	type fields struct {
		Scope  string
		Access Access
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "ok",
			fields: fields{
				Scope:  "a/b/c",
				Access: AccessCreate,
			},
			want: "a/b/c:" + strconv.Itoa(int(AccessCreate)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Permission{
				Scope:  tt.fields.Scope,
				Access: tt.fields.Access,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_HasPermission(t *testing.T) {
	tests := []struct {
		name string
		role Role
		args Permission
		want bool
	}{
		{
			name: "allowed permission => true",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/c", AccessRead}},
				DeniedPermissions:  nil,
			},
			args: Permission{"a/b/c", AccessRead},
			want: true,
		},
		{
			name: "allowed permission => false",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/c", AccessRead}},
				DeniedPermissions:  nil,
			},
			args: Permission{"a/b/c", AccessCreate},
			want: false,
		},
		{
			name: "same permission in allowed & denied => false",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/c", AccessRead}},
				DeniedPermissions:  []Permission{{"a/b/c", AccessRead}},
			},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "same permission in allowed & denied but denied has broader scope => false",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/c", AccessRead}},
				DeniedPermissions:  []Permission{{"a/b/*", AccessRead}},
			},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "same permission in allowed & denied but allowed has broader scope => false",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/*", AccessRead}},
				DeniedPermissions:  []Permission{{"a/b/c", AccessRead}},
			},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "same permission in allowed & denied but allowed has broader scope => true",
			role: Role{
				AllowedPermissions: []Permission{{"a/b/*", AccessRead}},
				DeniedPermissions:  []Permission{{"a/b/c", AccessRead}},
			},
			args: Permission{"a/b/d", AccessRead},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.HasPermission(tt.args); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_HasRole(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		args  string
		want  bool
	}{
		{
			name: "no roles => false",
			group: Group{
				Roles: nil,
			},
			args: "role1",
			want: false,
		},
		{
			name: "case sensitive => true",
			group: Group{
				Roles: []Role{{Name: "role1"}},
			},
			args: "role1",
			want: true,
		},
		{
			name: "case insensitive => false",
			group: Group{
				Roles: []Role{{Name: "role1"}},
			},
			args: "Role1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.group.HasRole(tt.args); got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_HasPermission(t *testing.T) {
	tests := []struct {
		name  string
		group Group
		args  Permission
		want  bool
	}{
		{
			name:  "group without roles => false",
			group: Group{},
			args:  Permission{"a/b/c", AccessRead},
			want:  false,
		},
		{
			name:  "group with 1 role without permissions => false",
			group: Group{Roles: []Role{{Name: "role1"}}},
			args:  Permission{"a/b/c", AccessRead},
			want:  false,
		},
		{
			name:  "group with 1 role and the role has the permission => true",
			group: Group{Roles: []Role{{AllowedPermissions: []Permission{{"a/b/c", AccessRead}}}}},
			args:  Permission{"a/b/c", AccessRead},
			want:  true,
		},
		{
			name:  "group with 2 roles and one role has the permission => true",
			group: Group{Roles: []Role{{AllowedPermissions: []Permission{{"a/b/d", AccessRead}, {"a/b/c", AccessRead}}}}},
			args:  Permission{"a/b/c", AccessRead},
			want:  true,
		},
		{
			name:  "group with 2 roles and none of them has the permission => false",
			group: Group{Roles: []Role{{AllowedPermissions: []Permission{{"a/b/d", AccessRead}, {"a/b/c", AccessRead}}}}},
			args:  Permission{"a/b/x", AccessRead},
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.group.HasPermission(tt.args); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Identity(t *testing.T) {
	tests := []struct {
		name string
		user User
		want string
	}{
		{
			name: "empty id",
			user: User{},
			want: "",
		},
		{
			name: "non empty id",
			user: User{Id: "123"},
			want: "123",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.Identity(); got != tt.want {
				t.Errorf("Identity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasRole(t *testing.T) {
	tests := []struct {
		name string
		user User
		args string
		want bool
	}{
		{
			name: "user without groups => false",
			user: User{},
			args: "role1",
			want: false,
		},
		{
			name: "user with 1 group and without roles => false",
			user: User{Groups: []Group{{}}},
			args: "role1",
			want: false,
		},
		{
			name: "user with 1 group and the role (case sensitive) => true",
			user: User{Groups: []Group{{Roles: []Role{{Name: "role1"}}}}},
			args: "role1",
			want: true,
		},
		{
			name: "user with 1 group and the role (case insensitive) => true",
			user: User{Groups: []Group{{Roles: []Role{{Name: "role1"}}}}},
			args: "Role1",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.HasRole(tt.args); got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_HasPermission(t *testing.T) {
	tests := []struct {
		name string
		user User
		args Permission
		want bool
	}{
		{
			name: "user without groups => false",
			user: User{},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "user with 1 group and without roles => false",
			user: User{Groups: []Group{{}}},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "user with 1 group, 1 role and without permissions => false",
			user: User{Groups: []Group{{Roles: []Role{{}}}}},
			args: Permission{"a/b/c", AccessRead},
			want: false,
		},
		{
			name: "user with 1 group, 1 role with permissions => true",
			user: User{Groups: []Group{{Roles: []Role{{AllowedPermissions: []Permission{{"a/b/c", AccessRead}}}}}}},
			args: Permission{"a/b/c", AccessRead},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.HasPermission(tt.args); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSecurityPrincipalFromContext(t *testing.T) {
	user := User{
		Id:               "123",
		Name:             "John",
		IdentityPlatform: "google",
	}
	tests := []struct {
		name string
		args context.Context
		want SecurityPrincipal
	}{
		{
			name: "empty context => nil",
			args: context.Background(),
			want: nil,
		},
		{
			name: "context with User => user",
			args: context.WithValue(context.Background(), KeyValues, user),
			want: user,
		},
		{
			name: "context with wrong type => nil",
			args: context.WithValue(context.Background(), KeyValues, "user"),
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSecurityPrincipalFromContext(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSecurityPrincipalFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
