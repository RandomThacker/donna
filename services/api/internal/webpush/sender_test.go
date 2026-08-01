package webpush

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

func TestPayloadFromNotification(t *testing.T) {
	t.Parallel()
	occ := "dev_abc_20260730T120000Z"
	typ := constant.NotificationTypeEvent
	n := entity.Notification{
		ID:               uuid.MustParse("018f0000-0000-7000-8000-000000000301"),
		Title:            "Standup",
		Body:             "Starts in 15 minutes",
		OccurrenceID:     &occ,
		NotificationType: &typ,
		Payload: json.RawMessage(`{
			"source":"DONNA",
			"type":"EVENT",
			"startAt":"2026-07-30T12:15:00Z",
			"occurrenceId":"dev_abc_20260730T120000Z",
			"deepLink":"/dashboard"
		}`),
	}
	p := PayloadFromNotification(n)
	if p.Title != "Standup" || p.Body == "" {
		t.Fatalf("payload = %+v", p)
	}
	if p.OccurrenceID != occ {
		t.Fatalf("occurrence = %s", p.OccurrenceID)
	}
	if p.DeepLink != constant.NotificationDeepLinkPath {
		t.Fatalf("deepLink = %s", p.DeepLink)
	}
	if p.Source != "DONNA" || p.TimelineType != "EVENT" {
		t.Fatalf("meta = %+v", p)
	}
	if p.StartTime != "2026-07-30T12:15:00Z" {
		t.Fatalf("start = %s", p.StartTime)
	}
}

func TestPayloadFallbackDeepLink(t *testing.T) {
	t.Parallel()
	scheduled := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	n := entity.Notification{
		ID:           uuid.MustParse("018f0000-0000-7000-8000-000000000302"),
		Title:        "x",
		Body:         "y",
		ScheduledFor: &scheduled,
	}
	p := PayloadFromNotification(n)
	if p.DeepLink != constant.NotificationDeepLinkPath {
		t.Fatalf("deepLink = %s", p.DeepLink)
	}
	if p.StartTime != scheduled.UTC().Format(time.RFC3339Nano) {
		t.Fatalf("start = %s", p.StartTime)
	}
}
