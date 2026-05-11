package redactor

import (
	"testing"

	"github.com/Di3Z1E/neuralog/internal/config"
)

// stubManager returns a config.Manager backed by a temp dir with the given config.
func stubManager(t *testing.T, cfg config.Config) *config.Manager {
	t.Helper()
	m := config.NewManager(t.TempDir())
	if err := m.Update(cfg); err != nil {
		t.Fatalf("stubManager: %v", err)
	}
	return m
}

func TestBuiltinPatterns(t *testing.T) {
	r := New(stubManager(t, config.Config{RedactEnabled: true}))

	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			"JWT",
			"token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			"[REDACTED:JWT]",
		},
		{
			"BearerToken",
			"Authorization: Bearer ghp_supersecrettoken1234567890abcdef",
			"[REDACTED:BEARER_TOKEN]",
		},
		{
			"AWSKeyID",
			"key=AKIAIOSFODNN7EXAMPLE access",
			"[REDACTED:AWS_KEY_ID]",
		},
		{
			"Password",
			"password=hunter2 connecting",
			"[REDACTED:PASSWORD]",
		},
		{
			"DatabaseURL",
			"connecting to postgres://user:s3cr3t@db.example.com/mydb",
			"[REDACTED:DB_URL]",
		},
		{
			"BasicAuthURL",
			"fetch https://user:pass@api.example.com/data",
			"[REDACTED:CREDENTIALS]@",
		},
		{
			"CreditCard",
			"card 4111111111111111 charged",
			"[REDACTED:CARD]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := r.Apply(tc.input)
			if got == tc.input {
				t.Errorf("Apply(%q): expected redaction, got unchanged line", tc.input)
			}
			if tc.want != "" {
				found := false
				for _, substr := range []string{tc.want} {
					if len(got) > 0 {
						found = containsSubstr(got, substr)
					}
				}
				if !found {
					t.Errorf("Apply(%q) = %q, want it to contain %q", tc.input, got, tc.want)
				}
			}
		})
	}
}

func TestDisabledRedactor(t *testing.T) {
	r := New(stubManager(t, config.Config{RedactEnabled: false}))
	line := "password=supersecret token=eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ4In0.abc"
	if got := r.Apply(line); got != line {
		t.Errorf("disabled Apply changed line: got %q", got)
	}
}

func TestCustomPatterns(t *testing.T) {
	cfg := config.Config{
		RedactEnabled: true,
		CustomPatterns: []config.RedactPattern{
			{ID: "1", Pattern: `txid-[0-9a-f]{8}`, Replace: "[REDACTED:TXN]"},
		},
	}
	r := New(stubManager(t, cfg))

	input := "processing txid-deadbeef done"
	got := r.Apply(input)
	if got == input {
		t.Error("custom pattern did not redact line")
	}
	if !containsSubstr(got, "[REDACTED:TXN]") {
		t.Errorf("Apply(%q) = %q, want [REDACTED:TXN]", input, got)
	}
}

func TestReload(t *testing.T) {
	mgr := stubManager(t, config.Config{RedactEnabled: true})
	r := New(mgr)

	// No custom pattern yet — line unchanged
	input := "txid-cafebabe hit"
	if got := r.Apply(input); got != input {
		t.Errorf("before reload: unexpected change: %q", got)
	}

	// Add custom pattern and reload
	if err := mgr.Update(config.Config{
		RedactEnabled: true,
		CustomPatterns: []config.RedactPattern{
			{ID: "1", Pattern: `txid-[0-9a-f]{8}`, Replace: "[REDACTED:TXN]"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	r.Reload()

	got := r.Apply(input)
	if got == input {
		t.Error("after reload: custom pattern not applied")
	}
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
