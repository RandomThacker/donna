package actions

import "github.com/RandomThacker/donna/services/api/internal/business"

// Registry holds constructed Actions for dependency injection into handlers.
type Registry struct {
	CreateEvent                 *CreateEventAction
	UpdateEvent                 *UpdateEventAction
	DeleteEvent                 *DeleteEventAction
	CreateReminder              *CreateReminderAction
	UpdateReminder              *UpdateReminderAction
	DeleteReminder              *DeleteReminderAction
	CreateAutomation            *CreateAutomationAction
	UpdateAutomation            *UpdateAutomationAction
	DeleteAutomation            *DeleteAutomationAction
	ListAutomations             *ListAutomationsAction
	ListAutomationTemplates     *ListAutomationTemplatesAction
	ListAutomationHistory       *ListAutomationHistoryAction
	ListAllAutomationHistory    *ListAllAutomationHistoryAction
	GetAutomationExecution      *GetAutomationExecutionAction
	GetAutomationAnalytics      *GetAutomationAnalyticsAction
	RunAutomation               *RunAutomationAction
	PreviewAutomation           *PreviewAutomationAction
	CreateTask                  *CreateTaskAction
	UpdateTask                  *UpdateTaskAction
	CompleteTask                *CompleteTaskAction
	DeleteTask                  *DeleteTaskAction
	QueryTimeline               *QueryTimelineAction
	GetNotifications            *GetNotificationsAction
	MarkNotificationRead        *MarkNotificationReadAction
	DismissNotification         *DismissNotificationAction
	ListDayTasks                *ListDayTasksAction
	GetPersonality              *GetPersonalityAction
	UpdatePersonality           *UpdatePersonalityAction
	ListPersonalityCatalog      *ListPersonalityCatalogAction
	PreviewPersonality          *PreviewPersonalityAction
}

// Deps wires services into the Action registry.
type Deps struct {
	Events               *business.DonnaEventService
	Reminders            *business.DonnaReminderService
	Automations          *business.AutomationService
	AutomationExecutions *business.AutomationExecutionService
	AutomationRunner     *business.AutomationRunner
	Tasks                *business.TaskJournalService
	Timeline             *business.TimelineService
	Notifications        *business.NotificationService
	Publisher            DomainEventPublisher
}

// NewRegistry constructs all Phase 2.4 Actions.
func NewRegistry(d Deps) *Registry {
	pub := d.Publisher
	if pub == nil {
		pub = NoopPublisher{}
	}
	return &Registry{
		CreateEvent:              NewCreateEventAction(d.Events, pub),
		UpdateEvent:              NewUpdateEventAction(d.Events, pub),
		DeleteEvent:              NewDeleteEventAction(d.Events, pub),
		CreateReminder:           NewCreateReminderAction(d.Reminders, pub),
		UpdateReminder:           NewUpdateReminderAction(d.Reminders, pub),
		DeleteReminder:           NewDeleteReminderAction(d.Reminders, pub),
		CreateAutomation:         NewCreateAutomationAction(d.Automations),
		UpdateAutomation:         NewUpdateAutomationAction(d.Automations),
		DeleteAutomation:         NewDeleteAutomationAction(d.Automations),
		ListAutomations:          NewListAutomationsAction(d.Automations, d.AutomationExecutions),
		ListAutomationTemplates:  NewListAutomationTemplatesAction(d.Automations),
		ListAutomationHistory:    NewListAutomationHistoryAction(d.AutomationExecutions),
		ListAllAutomationHistory: NewListAllAutomationHistoryAction(d.AutomationExecutions),
		GetAutomationExecution:   NewGetAutomationExecutionAction(d.AutomationExecutions),
		GetAutomationAnalytics:   NewGetAutomationAnalyticsAction(d.AutomationExecutions),
		RunAutomation:            NewRunAutomationAction(d.Automations, d.AutomationRunner),
		PreviewAutomation:        NewPreviewAutomationAction(d.Automations, d.AutomationRunner),
		CreateTask:               NewCreateTaskAction(d.Tasks, pub),
		UpdateTask:               NewUpdateTaskAction(d.Tasks, pub),
		CompleteTask:             NewCompleteTaskAction(d.Tasks, pub),
		DeleteTask:               NewDeleteTaskAction(d.Tasks, pub),
		QueryTimeline:            NewQueryTimelineAction(d.Timeline),
		GetNotifications:         NewGetNotificationsAction(d.Notifications),
		MarkNotificationRead:     NewMarkNotificationReadAction(d.Notifications, pub),
		DismissNotification:      NewDismissNotificationAction(d.Notifications, pub),
		ListDayTasks:             NewListDayTasksAction(&taskDayAdapter{svc: d.Tasks}),
	}
}
