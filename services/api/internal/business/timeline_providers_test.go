package business

import (
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

func TestMapCalendarEventToTimeline(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("018f0000-0000-7000-8000-000000000001")
	sourceID := uuid.MustParse("018f0000-0000-7000-8000-000000000002")
	desc := "standup notes"
	tz := "Asia/Kolkata"
	color := "#aabbcc"
	start := time.Date(2026, 7, 30, 10, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)

	item := mapCalendarEventToTimeline(entity.CalendarEventWithProvider{
		Event: entity.CalendarEvent{
			ID:               id,
			PublicID:         "evt_test",
			CalendarSourceID: sourceID,
			Title:            "Standup",
			Description:      &desc,
			StartsAt:         start,
			EndsAt:           end,
			IsAllDay:         false,
			Timezone:         &tz,
			Status:           constant.CalendarEventStatusConfirmed,
			Origin:           constant.CalendarEventOriginProviderSync,
		},
		Provider:    constant.AuthProviderGoogle,
		SourceColor: &color,
	})

	if item.Source != constant.TimelineSourceGoogle {
		t.Fatalf("source = %q, want GOOGLE", item.Source)
	}
	if item.Type != constant.TimelineTypeEvent {
		t.Fatalf("type = %q, want EVENT", item.Type)
	}
	if item.Status != constant.TimelineStatusActive {
		t.Fatalf("status = %q, want ACTIVE", item.Status)
	}
	if !item.ReadOnly {
		t.Fatal("provider events must be read-only")
	}
	if item.Color == nil || *item.Color != color {
		t.Fatalf("color = %v, want %s", item.Color, color)
	}
	if item.Metadata["provider_event_id"] != nil {
		t.Fatal("unexpected provider_event_id")
	}
}

func TestMapDonnaReminderToTimeline(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("018f0000-0000-7000-8000-000000000003")
	trigger := time.Date(2026, 7, 30, 15, 0, 0, 0, time.UTC)
	rule := "FREQ=DAILY"

	item := mapDonnaReminderToTimeline(entity.DonnaReminder{
		ID:             id,
		PublicID:       "drm_test",
		Title:          "Drink water",
		TriggerAt:      trigger,
		Timezone:       "UTC",
		RecurrenceRule: &rule,
		Status:         constant.DonnaReminderStatusScheduled,
	})

	if item.Source != constant.TimelineSourceDonna {
		t.Fatalf("source = %q, want DONNA", item.Source)
	}
	if item.Type != constant.TimelineTypeReminder {
		t.Fatalf("type = %q, want REMINDER", item.Type)
	}
	if item.Status != constant.TimelineStatusActive {
		t.Fatalf("status = %q, want ACTIVE", item.Status)
	}
	if item.ReadOnly {
		t.Fatal("donna reminders must be writable")
	}
	if !item.StartAt.Equal(trigger) || !item.EndAt.Equal(trigger) {
		t.Fatalf("reminder window = %v..%v, want %v", item.StartAt, item.EndAt, trigger)
	}
	// mapDonnaReminderToTimeline is the non-expanded mapper used only as a helper;
	// expansion sets IsRecurring. Direct map of a row with a rule still marks metadata.
	if item.Metadata["public_id"] != "drm_test" {
		t.Fatalf("metadata = %v", item.Metadata)
	}
}

func TestTimelineSourceFromProvider(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		constant.AuthProviderGoogle:    constant.TimelineSourceGoogle,
		constant.AuthProviderMicrosoft: constant.TimelineSourceMicrosoftICS,
		constant.AuthProviderICS:       constant.TimelineSourceMicrosoftICS,
	}
	for provider, want := range cases {
		if got := timelineSourceFromProvider(provider); got != want {
			t.Fatalf("provider %q → %q, want %q", provider, got, want)
		}
	}
}

func TestTimelineStatusCancelled(t *testing.T) {
	t.Parallel()
	if got := timelineStatusFromCalendar(constant.CalendarEventStatusCancelled); got != constant.TimelineStatusCancelled {
		t.Fatalf("calendar cancelled → %q", got)
	}
	if got := timelineStatusFromDonnaReminder(constant.DonnaReminderStatusCancelled); got != constant.TimelineStatusCancelled {
		t.Fatalf("reminder cancelled → %q", got)
	}
}
