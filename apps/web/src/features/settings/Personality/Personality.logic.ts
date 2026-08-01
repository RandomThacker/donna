export function defaultBrowserTimezone(): string {
  try {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
    if (tz && tz.trim()) {
      return tz.trim();
    }
  } catch {
    // fall through
  }
  return "UTC";
}

export const EMOJI_LEVELS = ["none", "low", "medium", "high"] as const;

export const PREVIEW_LABELS: Record<string, string> = {
  greeting: "Morning Greeting",
  reminder: "Reminder",
  task_complete: "Task Completion",
  error: "Error Message",
  notification: "Notification",
  automation: "Automation Summary",
  morning_brief: "Morning Brief",
  chat: "Chat Response",
};
