package actions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/RandomThacker/donna/services/api/internal/actions"
	"github.com/RandomThacker/donna/services/api/internal/apperr"
	"github.com/RandomThacker/donna/services/api/internal/business"
	"github.com/RandomThacker/donna/services/api/internal/entity"
	"github.com/google/uuid"
)

type recordingPublisher struct {
	events []actions.DomainEvent
}

func (p *recordingPublisher) Publish(_ context.Context, event actions.DomainEvent) error {
	p.events = append(p.events, event)
	return nil
}

type stubEventService struct {
	createIn  business.CreateDonnaEventInput
	createOut entity.DonnaEvent
	createErr error
	updateIn  business.UpdateDonnaEventInput
	updateOut entity.DonnaEvent
	updateErr error
	deleteErr error
	deletedID uuid.UUID
}

func (s *stubEventService) Create(_ context.Context, userID uuid.UUID, in business.CreateDonnaEventInput) (entity.DonnaEvent, error) {
	s.createIn = in
	if s.createErr != nil {
		return entity.DonnaEvent{}, s.createErr
	}
	out := s.createOut
	out.UserID = userID
	if out.ID == uuid.Nil {
		out.ID = uuid.MustParse("018f0000-0000-7000-8000-000000000401")
	}
	return out, nil
}

func (s *stubEventService) Update(_ context.Context, _, _ uuid.UUID, in business.UpdateDonnaEventInput) (entity.DonnaEvent, error) {
	s.updateIn = in
	if s.updateErr != nil {
		return entity.DonnaEvent{}, s.updateErr
	}
	return s.updateOut, nil
}

func (s *stubEventService) Delete(_ context.Context, _, eventID uuid.UUID) error {
	s.deletedID = eventID
	return s.deleteErr
}

func TestCreateEventActionOrchestration(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000402")
	start := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	svc := &stubEventService{
		createOut: entity.DonnaEvent{
			PublicID: "dev_1", Title: "Guitar", StartAt: start, EndAt: start.Add(time.Hour),
			Timezone: "UTC", Status: "CONFIRMED",
			CreatedAt: start, UpdatedAt: start,
		},
	}
	pub := &recordingPublisher{}
	action := actions.NewCreateEventAction(svc, pub)

	result, err := action.Execute(context.Background(), actions.CreateEventRequest{
		UserID: userID, Title: "Guitar", StartAt: start, EndAt: start.Add(time.Hour), Timezone: "UTC",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != "Guitar" || result.UserID != userID {
		t.Fatalf("result = %+v", result)
	}
	if svc.createIn.Title != "Guitar" {
		t.Fatalf("service input = %+v", svc.createIn)
	}
	if len(pub.events) != 1 || pub.events[0].Name != "event.created" {
		t.Fatalf("publisher = %+v", pub.events)
	}
}

func TestCreateEventActionValidation(t *testing.T) {
	t.Parallel()
	action := actions.NewCreateEventAction(&stubEventService{}, nil)
	_, err := action.Execute(context.Background(), actions.CreateEventRequest{Title: "x"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdateEventActionPassesPatch(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000403")
	eventID := uuid.MustParse("018f0000-0000-7000-8000-000000000404")
	title := "Renamed"
	svc := &stubEventService{
		updateOut: entity.DonnaEvent{
			ID: eventID, UserID: userID, PublicID: "dev_2", Title: title,
			StartAt: time.Now(), EndAt: time.Now().Add(time.Hour), Timezone: "UTC", Status: "CONFIRMED",
		},
	}
	action := actions.NewUpdateEventAction(svc, nil)
	result, err := action.Execute(context.Background(), actions.UpdateEventRequest{
		UserID: userID, EventID: eventID, Title: &title,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Title != title || svc.updateIn.Title == nil || *svc.updateIn.Title != title {
		t.Fatalf("result=%+v in=%+v", result, svc.updateIn)
	}
}

func TestDeleteEventAction(t *testing.T) {
	t.Parallel()
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000405")
	eventID := uuid.MustParse("018f0000-0000-7000-8000-000000000406")
	svc := &stubEventService{}
	action := actions.NewDeleteEventAction(svc, nil)
	if err := action.Execute(context.Background(), actions.DeleteEventRequest{UserID: userID, EventID: eventID}); err != nil {
		t.Fatal(err)
	}
	if svc.deletedID != eventID {
		t.Fatalf("deleted = %s", svc.deletedID)
	}
}

type stubTimelineService struct {
	items []entity.TimelineItem
	err   error
	from  time.Time
	to    time.Time
}

func (s *stubTimelineService) List(_ context.Context, _ uuid.UUID, from, to time.Time) ([]entity.TimelineItem, error) {
	s.from, s.to = from, to
	return s.items, s.err
}

func TestQueryTimelineActionValidation(t *testing.T) {
	t.Parallel()
	action := actions.NewQueryTimelineAction(&stubTimelineService{})
	_, err := action.Execute(context.Background(), actions.QueryTimelineRequest{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000407"),
		From:   time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}

func TestQueryTimelineActionOrchestration(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	svc := &stubTimelineService{
		items: []entity.TimelineItem{{ID: "e1", Title: "Meet", StartAt: from, EndAt: from.Add(time.Hour)}},
	}
	action := actions.NewQueryTimelineAction(svc)
	result, err := action.Execute(context.Background(), actions.QueryTimelineRequest{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000408"),
		From:   from,
		To:     to,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || result.Items[0].Title != "Meet" {
		t.Fatalf("result = %+v", result)
	}
	if !svc.from.Equal(from) || !svc.to.Equal(to) {
		t.Fatalf("range = %v %v", svc.from, svc.to)
	}
}

type stubNotificationService struct {
	listed     []entity.Notification
	listErr    error
	readOut    entity.Notification
	readErr    error
	dismissOut entity.Notification
	dismissErr error
	statuses   []string
}

func (s *stubNotificationService) List(_ context.Context, _ uuid.UUID, statuses []string) ([]entity.Notification, error) {
	s.statuses = statuses
	return s.listed, s.listErr
}

func (s *stubNotificationService) MarkRead(_ context.Context, _, _ uuid.UUID) (entity.Notification, error) {
	return s.readOut, s.readErr
}

func (s *stubNotificationService) MarkDismissed(_ context.Context, _, _ uuid.UUID) (entity.Notification, error) {
	return s.dismissOut, s.dismissErr
}

func TestGetNotificationsAction(t *testing.T) {
	t.Parallel()
	svc := &stubNotificationService{
		listed: []entity.Notification{{
			ID: uuid.MustParse("018f0000-0000-7000-8000-000000000409"),
			Title: "t", Body: "b", Status: "PENDING",
		}},
	}
	action := actions.NewGetNotificationsAction(svc)
	out, err := action.Execute(context.Background(), actions.GetNotificationsRequest{
		UserID:   uuid.MustParse("018f0000-0000-7000-8000-000000000410"),
		Statuses: []string{"PENDING"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Title != "t" {
		t.Fatalf("out = %+v", out)
	}
	if len(svc.statuses) != 1 || svc.statuses[0] != "PENDING" {
		t.Fatalf("statuses = %v", svc.statuses)
	}
}

func TestMarkNotificationReadAction(t *testing.T) {
	t.Parallel()
	id := uuid.MustParse("018f0000-0000-7000-8000-000000000411")
	svc := &stubNotificationService{
		readOut: entity.Notification{ID: id, Title: "t", Body: "b", Status: "READ"},
	}
	action := actions.NewMarkNotificationReadAction(svc, nil)
	out, err := action.Execute(context.Background(), actions.MarkNotificationReadRequest{
		UserID: uuid.MustParse("018f0000-0000-7000-8000-000000000412"), NotificationID: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "READ" {
		t.Fatalf("status = %s", out.Status)
	}
}

type stubTaskService struct {
	createOut entity.TaskOccurrenceWithTask
	createErr error
	updateOut entity.Task
	updateErr error
	tagsOut   []entity.TaskTag
	tagsErr   error
	occOut    entity.TaskOccurrenceWithTask
	occErr    error
	deleteErr error
	completed *bool
}

func (s *stubTaskService) CreateTask(context.Context, uuid.UUID, business.CreateTaskInput) (entity.TaskOccurrenceWithTask, error) {
	return s.createOut, s.createErr
}

func (s *stubTaskService) UpdateTask(context.Context, uuid.UUID, uuid.UUID, business.UpdateTaskInput) (entity.Task, error) {
	return s.updateOut, s.updateErr
}

func (s *stubTaskService) UpdateOccurrence(_ context.Context, _, _ uuid.UUID, completed bool) (entity.TaskOccurrenceWithTask, error) {
	s.completed = &completed
	return s.occOut, s.occErr
}

func (s *stubTaskService) DeleteTask(context.Context, uuid.UUID, uuid.UUID) error {
	return s.deleteErr
}

func (s *stubTaskService) ListTaskTagsForTask(context.Context, uuid.UUID, uuid.UUID) ([]entity.TaskTag, error) {
	return s.tagsOut, s.tagsErr
}

func TestUpdateTaskActionLoadsTags(t *testing.T) {
	t.Parallel()
	taskID := uuid.MustParse("018f0000-0000-7000-8000-000000000413")
	userID := uuid.MustParse("018f0000-0000-7000-8000-000000000414")
	svc := &stubTaskService{
		updateOut: entity.Task{ID: taskID, UserID: userID, Title: "Write", UpdatedAt: time.Now()},
		tagsOut:   []entity.TaskTag{{ID: uuid.MustParse("018f0000-0000-7000-8000-000000000415"), Name: "focus", Color: "#fff"}},
	}
	action := actions.NewUpdateTaskAction(svc, nil)
	out, err := action.Execute(context.Background(), actions.UpdateTaskRequest{UserID: userID, TaskID: taskID})
	if err != nil {
		t.Fatal(err)
	}
	if out.Title != "Write" || len(out.Tags) != 1 || out.Tags[0].Name != "focus" {
		t.Fatalf("out = %+v", out)
	}
}

func TestCompleteTaskAction(t *testing.T) {
	t.Parallel()
	svc := &stubTaskService{
		occOut: entity.TaskOccurrenceWithTask{
			TaskOccurrence: entity.TaskOccurrence{
				ID: uuid.MustParse("018f0000-0000-7000-8000-000000000416"),
				Completed: true,
			},
			Title: "Done",
		},
	}
	action := actions.NewCompleteTaskAction(svc, nil)
	out, err := action.Execute(context.Background(), actions.CompleteTaskRequest{
		UserID:       uuid.MustParse("018f0000-0000-7000-8000-000000000417"),
		OccurrenceID: uuid.MustParse("018f0000-0000-7000-8000-000000000416"),
		Completed:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Completed || svc.completed == nil || !*svc.completed {
		t.Fatalf("out=%+v completed=%v", out, svc.completed)
	}
}

func TestCreateTaskActionValidation(t *testing.T) {
	t.Parallel()
	action := actions.NewCreateTaskAction(&stubTaskService{}, nil)
	_, err := action.Execute(context.Background(), actions.CreateTaskRequest{Title: "x"})
	if !errors.Is(err, apperr.ErrValidation) {
		t.Fatalf("err = %v", err)
	}
}
