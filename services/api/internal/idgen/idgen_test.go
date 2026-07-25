package idgen_test

import (
	"strings"
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
)

func TestNewUUIDv7AndPublicID(t *testing.T) {
	id, err := idgen.NewUUIDv7()
	if err != nil {
		t.Fatalf("NewUUIDv7: %v", err)
	}
	publicID := idgen.PublicID(constant.PublicIDPrefixUser, id)
	if !strings.HasPrefix(publicID, constant.PublicIDPrefixUser) {
		t.Fatalf("public_id = %q", publicID)
	}
	if strings.Contains(publicID, "-") {
		t.Fatalf("public_id should omit dashes: %q", publicID)
	}
	if len(publicID) != len(constant.PublicIDPrefixUser)+32 {
		t.Fatalf("public_id length = %d (%q)", len(publicID), publicID)
	}
}
