export type NotificationStatus =
  | "PENDING"
  | "SENT"
  | "READ"
  | "DISMISSED"
  | "FAILED";

export type NotificationType = "EVENT" | "REMINDER";

export type NotificationFilter =
  | "all"
  | "unread"
  | "pending"
  | "sent"
  | "failed"
  | "dismissed";

export type NotificationPayload = {
  timelineItemId?: string;
  occurrenceId?: string;
  source?: string;
  type?: string;
  startAt?: string;
  endAt?: string;
  timezone?: string;
  deepLink?: string;
  parentId?: string;
  recurrenceRule?: string;
  failureReason?: string;
  error?: string;
};

export type DonnaNotification = {
  id: string;
  public_id: string;
  timeline_item_parent_id?: string | null;
  occurrence_id?: string | null;
  title: string;
  body: string;
  notification_type?: NotificationType | string | null;
  scheduled_for?: string | null;
  status: NotificationStatus | string;
  delivery_channels: string[];
  channel_delivery_status?: Record<string, string> | null;
  payload?: NotificationPayload | Record<string, unknown> | null;
  read_at?: string | null;
  dismissed_at?: string | null;
  sent_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type NotificationsListResponse = {
  notifications: DonnaNotification[];
};

export type NotificationGroupKey =
  | "today"
  | "yesterday"
  | "earlier_this_week"
  | "older";

export type NotificationGroup = {
  key: NotificationGroupKey;
  label: string;
  items: DonnaNotification[];
};

export type StatusTimelineStep = {
  id: string;
  label: string;
  at: string | null;
  done: boolean;
};
