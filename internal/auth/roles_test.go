package auth

import "testing"

func TestAtLeast(t *testing.T) {
	cases := []struct {
		have, need string
		want       bool
	}{
		{RoleAdmin, RoleAdmin, true},
		{RoleAdmin, RoleMember, true},
		{RoleAdmin, RoleViewer, true},
		{RoleMember, RoleViewer, true},
		{RoleMember, RoleMember, true},
		{RoleMember, RoleAdmin, false},
		{RoleViewer, RoleMember, false},
		{RoleViewer, RoleViewer, true},
		{"root", RoleViewer, false},
		{RoleAdmin, "root", false},
		{"", RoleViewer, false},
	}
	for _, tc := range cases {
		if got := AtLeast(tc.have, tc.need); got != tc.want {
			t.Errorf("AtLeast(%q, %q) = %v, want %v", tc.have, tc.need, got, tc.want)
		}
	}
}

func TestValidRole(t *testing.T) {
	for _, r := range []string{RoleViewer, RoleMember, RoleAdmin} {
		if !ValidRole(r) {
			t.Errorf("ValidRole(%q) = false", r)
		}
	}
	for _, r := range []string{"", "root", "Admin", "member "} {
		if ValidRole(r) {
			t.Errorf("ValidRole(%q) = true, want false", r)
		}
	}
}
