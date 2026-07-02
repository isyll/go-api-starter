package env_test

import (
	"testing"

	"github.com/isyll/go-grpc-starter/pkg/env"
)

func TestGetOrDefault(t *testing.T) {
	t.Run("env var set", func(t *testing.T) {
		t.Setenv("TEST_GET_ENV_KEY", "hello world")
		got := env.GetOrDefault("TEST_GET_ENV_KEY", "default")
		// sanitize lowercases and collapses; "hello world" stays.
		if got != "hello world" {
			t.Fatalf("got %q, want %q", got, "hello world")
		}
	})
	t.Run("env var unset", func(t *testing.T) {
		got := env.GetOrDefault("TEST_GET_ENV_MISSING", "fallback")
		if got != "fallback" {
			t.Fatalf("got %q, want %q", got, "fallback")
		}
	})
	t.Run("env var set to empty", func(t *testing.T) {
		t.Setenv("TEST_GET_ENV_EMPTY", "")
		got := env.GetOrDefault("TEST_GET_ENV_EMPTY", "default")
		if got != "default" {
			t.Fatalf("got %q, want %q", got, "default")
		}
	})
}

func TestIsDev(t *testing.T) {
	cases := []struct {
		name string
		set  string
		want bool
	}{
		{"development", env.Development, true},
		{"dev short", "dev", true},
		{"testing", env.Testing, true},
		{"test short", "test", true},
		{"production", env.Production, false},
		{"prod short", "prod", false},
		// IsDev is a strict predicate; defaulting to development
		// when both APP_ENV and GO_ENV are unset is InitApp's job,
		// not IsDev's.
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("APP_ENV", tc.set)
			t.Setenv("GO_ENV", "")
			if got := env.IsDev(); got != tc.want {
				t.Fatalf("IsDev()=%v, want %v", got, tc.want)
			}
		})
	}
}
