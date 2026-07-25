package oauthstate_test

import (
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/oauthstate"
)

func TestStateRoundTrip(t *testing.T) {
	m := oauthstate.NewManager("secret")
	state, err := m.Create()
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Verify(state); err != nil {
		t.Fatal(err)
	}
	if err := m.Verify(state + "x"); err == nil {
		t.Fatal("expected verify failure")
	}
}
