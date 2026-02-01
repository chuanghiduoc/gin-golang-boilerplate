package entity

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func (r Role) String() string {
	return string(r)
}

func (r Role) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin:
		return true
	}
	return false
}

func (r Role) IsAdmin() bool {
	return r == RoleAdmin
}

func ParseRole(s string) Role {
	switch s {
	case "admin":
		return RoleAdmin
	default:
		return RoleUser
	}
}

var AllRoles = []Role{RoleUser, RoleAdmin}
