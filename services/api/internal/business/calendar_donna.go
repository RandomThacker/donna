package business

import (
	"time"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/RandomThacker/donna/services/api/internal/idgen"
	"github.com/google/uuid"
)

// DonnaCalendarSourceID returns a stable virtual calendar source id for a user.
// It is never persisted — used so the calendar UI can show Donna-owned events.
func DonnaCalendarSourceID(userID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte("donna://calendar/source/"+userID.String()))
}

func donnaCalendarSource(userID uuid.UUID, now time.Time) entity.CalendarSource {
	id := DonnaCalendarSourceID(userID)
	color := "#C9A87C"
	tz := constant.DefaultUserTimezone
	return entity.CalendarSource{
		ID:                 id,
		PublicID:           idgen.PublicID(constant.PublicIDPrefixCalendarSource, id),
		UserID:             userID,
		ConnectedAccountID: uuid.Nil,
		ProviderCalendarID: constant.CalendarProviderCalendarIDDonna,
		Name:               constant.CalendarDonnaSourceName,
		Color:              &color,
		IsWritable:         true,
		SyncEnabled:        true,
		Timezone:           &tz,
		ProviderMetadata:   []byte(`{"virtual":true}`),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func mapDonnaEventToCalendarEvent(e entity.DonnaEvent, sourceID uuid.UUID) entity.CalendarEvent {
	tz := e.Timezone
	status := e.Status
	if status == "" {
		status = constant.CalendarEventStatusConfirmed
	}
	return entity.CalendarEvent{
		ID:               e.ID,
		PublicID:         e.PublicID,
		UserID:           e.UserID,
		CalendarSourceID: sourceID,
		Title:            e.Title,
		Description:      e.Description,
		Location:         e.Location,
		StartsAt:         e.StartAt,
		EndsAt:           e.EndAt,
		IsAllDay:         e.AllDay,
		Status:           status,
		Timezone:         &tz,
		RecurrenceRule:   e.RecurrenceRule,
		Origin:           constant.CalendarEventOriginDonna,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}
