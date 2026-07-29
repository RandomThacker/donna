export type TaskOccurrenceSource =
  | "manual"
  | "recurring"
  | "calendar"
  | "ai"
  | "carry_forward";

export type TaskTag = {
  id: string;
  public_id: string;
  name: string;
  color: string;
  updated_at?: string;
};

export type TaskOccurrence = {
  id: string;
  public_id: string;
  task_id: string;
  date: string;
  sort_order: number;
  completed: boolean;
  completed_at?: string | null;
  carried_forward: boolean;
  source: TaskOccurrenceSource;
  title: string;
  description?: string | null;
  priority?: string | null;
  project?: string | null;
  labels?: string[];
  tags?: TaskTag[];
  recurrence_rule?: string | null;
};

export type TaskDayStatistics = {
  total: number;
  completed: number;
  pending: number;
  carried: number;
  completion_pct: number;
  completed_today: number;
  carried_forward: number;
  longest_carried_streak: number;
  average_completion_min?: number | null;
  streak: number;
};

export type TaskDayResponse = {
  date: string;
  note: { content: string; updated_at?: string };
  statistics: TaskDayStatistics;
  occurrences: TaskOccurrence[];
};

export type TaskHistoryDay = {
  date: string;
  total: number;
  completed: number;
  pending: number;
  carried: number;
};

export type TaskHistoryResponse = {
  days: TaskHistoryDay[];
};

export type ReorderTaskOccurrencesInput = {
  date: string;
  occurrence_ids: string[];
};
