package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Payments":           "payments",
		"Payments API":       "payments-api",
		"  Hello,  World!  ": "hello-world",
		"API v2 -- (beta)":   "api-v2-beta",
		"":                   "",
		"---":                "",
		"Node_Service.exe":   "node-service-exe",
		"Über Service":       "ber-service",
		"a/b/c":              "a-b-c",
	}
	for in, want := range cases {
		if got := Make(in); got != want {
			t.Errorf("Make(%q) = %q, want %q", in, got, want)
		}
	}
}
