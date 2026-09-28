package auth

// Workspace roles. Higher rank implies all lower-rank permissions.
const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
)

func rank(role string) int {
	switch role {
	case RoleViewer:
		return 0
	case RoleMember:
		return 1
	case RoleAdmin:
		return 2
	default:
		return -1
	}
}

// ValidRole reports whether role is a known workspace role.
func ValidRole(role string) bool { return rank(role) >= 0 }

// AtLeast reports whether have is a known role with at least need's rank.
func AtLeast(have, need string) bool {
	hr, nr := rank(have), rank(need)
	return hr >= 0 && nr >= 0 && hr >= nr
}
