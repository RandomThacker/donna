package config

import (
	"testing"
)

func TestExpandEnv(t *testing.T) {
	t.Setenv("DONNA_TEST_A", "alpha")
	t.Setenv("DONNA_TEST_B", "")

	got := expandEnv(`x=${DONNA_TEST_A} y=${DONNA_TEST_B:fallback} z=${DONNA_TEST_MISSING:http://localhost:3000} p=${DONNA_TEST_PORT::8080}`)
	want := `x=alpha y=fallback z=http://localhost:3000 p=:8080`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
