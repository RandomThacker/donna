package automationcatalog_test

import (
	"testing"

	"github.com/RandomThacker/donna/services/api/internal/automationcatalog"
)

func TestCatalogLoadsTemplates(t *testing.T) {
	t.Parallel()
	templates, err := automationcatalog.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(templates) < 5 {
		t.Fatalf("expected starter templates, got %d", len(templates))
	}
	morning, ok := automationcatalog.Get("morning_brief")
	if !ok {
		t.Fatal("missing morning_brief")
	}
	if len(morning.Commands) < 2 {
		t.Fatalf("morning brief should be multi-command, got %#v", morning.Commands)
	}
	if morning.DefaultSchedule.Type != "daily" {
		t.Fatalf("schedule type = %s", morning.DefaultSchedule.Type)
	}
}
