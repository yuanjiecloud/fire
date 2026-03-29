package task

import "testing"

// ---------------------------------------------------------------------------
// Version
// ---------------------------------------------------------------------------

func TestVersion_String(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{"v1.0.0", "v1.0.0"},
		{"", ""},
		{"main", "main"},
	}
	for _, tc := range tests {
		t.Run(string(tc.v), func(t *testing.T) {
			if got := tc.v.String(); got != tc.want {
				t.Errorf("Version(%q).String() = %q, want %q", tc.v, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Replacement.IsLocal
// ---------------------------------------------------------------------------

func TestReplacement_IsLocal(t *testing.T) {
	tests := []struct {
		name       string
		repository string
		wantLocal  bool
	}{
		{"local relative path", "./my-repo", true},
		{"local absolute path", "/home/user/my-repo", true},
		{"bare name", "my-repo", true},
		{"http URL", "http://github.com/org/repo.git", false},
		{"https URL", "https://github.com/org/repo.git", false},
		{"git SSH", "git@github.com:org/repo.git", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Replacement{Repository: tc.repository}
			if got := r.IsLocal(); got != tc.wantLocal {
				t.Errorf("IsLocal() = %v, want %v (repo=%q)", got, tc.wantLocal, tc.repository)
			}
		})
	}
}
