package business

import (
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

func mustParseUUID(t *testing.T, raw string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func donnaEventFixture(id uuid.UUID, start, end time.Time, rule *string) entity.DonnaEvent {
	return entity.DonnaEvent{
		ID:             id,
		PublicID:       "dev_test",
		Title:          "Guitar Class",
		StartAt:        start,
		EndAt:          end,
		Timezone:       "UTC",
		Status:         constant.DonnaEventStatusConfirmed,
		RecurrenceRule: rule,
	}
}
