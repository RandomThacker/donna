export type TimelineView = "day" | "week" | "month" | "agenda";

export type TimelineSource = "GOOGLE" | "MICROSOFT_ICS" | "DONNA";
export type TimelineItemType = "EVENT" | "REMINDER";
export type TimelineStatus =
  | "ACTIVE"
  | "COMPLETED"
  | "CANCELLED"
  | "MISSED";

export type TimelineItem = {
  id: string;
  source: TimelineSource | string;
  type: TimelineItemType | string;
  status: TimelineStatus | string;
  title: string;
  description?: string | null;
  start_at: string;
  end_at: string;
  timezone: string;
  all_day: boolean;
  color?: string | null;
  read_only: boolean;
  metadata?: Record<string, unknown> | null;
  is_recurring: boolean;
  recurrence_rule?: string | null;
  parent_id?: string | null;
  occurrence_id: string;
  occurrence_start?: string | null;
  occurrence_end?: string | null;
};

export type TimelineResponse = {
  items: TimelineItem[];
  from: string;
  to: string;
};

export type CreateDonnaEventInput = {
  title: string;
  description?: string | null;
  start_at: string;
  end_at: string;
  timezone: string;
  all_day?: boolean;
  location?: string | null;
  reminder_offset_minutes?: number | null;
  recurrence_rule?: string | null;
  color?: string | null;
};

export type UpdateDonnaEventInput = {
  title?: string;
  description?: string | null;
  start_at?: string;
  end_at?: string;
  timezone?: string;
  all_day?: boolean;
  location?: string | null;
  reminder_offset_minutes?: number | null;
  recurrence_rule?: string | null;
  status?: string;
  color?: string | null;
};

export type DonnaEventResponse = {
  id: string;
  public_id: string;
  title: string;
  description?: string | null;
  start_at: string;
  end_at: string;
  timezone: string;
  all_day: boolean;
  location?: string | null;
  reminder_offset_minutes?: number | null;
  recurrence_rule?: string | null;
  status: string;
  color?: string | null;
  created_at: string;
  updated_at: string;
};

export type CreateDonnaReminderInput = {
  title: string;
  description?: string | null;
  trigger_at: string;
  timezone: string;
  recurrence_rule?: string | null;
  color?: string | null;
};

export type UpdateDonnaReminderInput = {
  title?: string;
  description?: string | null;
  trigger_at?: string;
  timezone?: string;
  recurrence_rule?: string | null;
  status?: string;
  color?: string | null;
};

export type DonnaReminderResponse = {
  id: string;
  public_id: string;
  title: string;
  description?: string | null;
  trigger_at: string;
  timezone: string;
  recurrence_rule?: string | null;
  status: string;
  color?: string | null;
  created_at: string;
  updated_at: string;
};

export type LaidOutTimelineItem = {
  item: TimelineItem;
  startMin: number;
  endMin: number;
  column: number;
  columns: number;
  top: number;
  height: number;
};
