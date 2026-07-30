package actions

import "github.com/RandomThacker/donna/services/api/internal/business"

// Registry holds constructed Actions for dependency injection into handlers.
type Registry struct {
	CreateEvent            *CreateEventAction
	UpdateEvent            *UpdateEventAction
	DeleteEvent            *DeleteEventAction
	CreateReminder         *CreateReminderAction
	UpdateReminder         *UpdateReminderAction
	DeleteReminder         *DeleteReminderAction
	CreateTask             *CreateTaskAction
	UpdateTask             *UpdateTaskAction
	CompleteTask           *CompleteTaskAction
	DeleteTask             *DeleteTaskAction
	QueryTimeline          *QueryTimelineAction
	GetNotifications       *GetNotificationsAction
	MarkNotificationRead   *MarkNotificationReadAction
	DismissNotification    *DismissNotificationAction
	ListDayTasks           *ListDayTasksAction
}

// Deps wires services into the Action registry.
type Deps struct {
	Events        *business.DonnaEventService
	Reminders     *business.DonnaReminderService
	Tasks         *business.TaskJournalService
	Timeline      *business.TimelineService
	Notifications *business.NotificationService
	Publisher     DomainEventPublisher
}

// NewRegistry constructs all Phase 2.4 Actions.
func NewRegistry(d Deps) *Registry {
	pub := d.Publisher
	if pub == nil {
		pub = NoopPublisher{}
	}
	return &Registry{
		CreateEvent:          NewCreateEventAction(d.Events, pub),
		UpdateEvent:          NewUpdateEventAction(d.Events, pub),
		DeleteEvent:          NewDeleteEventAction(d.Events, pub),
		CreateReminder:       NewCreateReminderAction(d.Reminders, pub),
		UpdateReminder:       NewUpdateReminderAction(d.Reminders, pub),
		DeleteReminder:       NewDeleteReminderAction(d.Reminders, pub),
		CreateTask:           NewCreateTaskAction(d.Tasks, pub),
		UpdateTask:           NewUpdateTaskAction(d.Tasks, pub),
		CompleteTask:         NewCompleteTaskAction(d.Tasks, pub),
		DeleteTask:           NewDeleteTaskAction(d.Tasks, pub),
		QueryTimeline:        NewQueryTimelineAction(d.Timeline),
		GetNotifications:     NewGetNotificationsAction(d.Notifications),
		MarkNotificationRead: NewMarkNotificationReadAction(d.Notifications, pub),
		DismissNotification:  NewDismissNotificationAction(d.Notifications, pub),
		ListDayTasks:         NewListDayTasksAction(&taskDayAdapter{svc: d.Tasks}),
	}
}
