import type {
  Automation,
  AutomationCommand,
  AutomationTemplate,
} from "./Automations.types";

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

export function formatLocalTimeForInput(localTime: string): string {
  const trimmed = localTime.trim();
  if (/^\d{2}:\d{2}/.test(trimmed)) {
    return trimmed.slice(0, 5);
  }
  return "09:00";
}

export function formatSchedule(auto: Automation): string {
  const time = formatLocalTimeForInput(auto.trigger?.time ?? "09:00");
  const type = auto.trigger?.type === "daily" ? "Daily" : auto.trigger?.type ?? "Daily";
  return `${type} at ${time}`;
}

export function formatDelivery(auto: Automation): string {
  const channels = auto.delivery?.channels ?? ["chat"];
  if (channels.length === 0) return "Donna Chat";
  return channels
    .map((ch) => {
      if (ch === "chat") return "Donna Chat";
      return ch.charAt(0).toUpperCase() + ch.slice(1);
    })
    .join(", ");
}

export function formatRunAt(iso: string | null | undefined): string {
  if (!iso) return "—";
  try {
    return new Date(iso).toLocaleString(undefined, {
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "2-digit",
    });
  } catch {
    return "—";
  }
}

export function commandLabel(command: AutomationCommand | string): string {
  if (typeof command === "string") {
    return command;
  }
  if (command.label?.trim()) {
    return command.label.trim();
  }
  if (command.command === "chat_message") {
    return command.variables?.message?.trim() || "Chat message";
  }
  switch (command.command) {
    case "greeting":
      return "Greeting";
    case "todays_agenda":
      return command.variables?.range === "tomorrow"
        ? "Tomorrow's Agenda"
        : "Today's Agenda";
    case "tasks_due":
      return "Tasks Due";
    default:
      return command.command || "Command";
  }
}

export function commandsPreview(
  commands: Array<AutomationCommand | string>,
): string {
  if (commands.length === 0) return "No commands";
  if (commands.length === 1) return commandLabel(commands[0]!);
  return `${commands.length} commands`;
}

export function formatStatusLabel(status: string | null | undefined): string {
  switch (status) {
    case "SUCCESS":
      return "Success";
    case "PARTIAL_SUCCESS":
      return "Partial";
    case "FAILED":
      return "Failed";
    case "RUNNING":
      return "Running";
    case "CANCELLED":
      return "Cancelled";
    default:
      return status?.trim() || "—";
  }
}

export function formatDurationMs(ms: number | null | undefined): string {
  if (ms == null || Number.isNaN(ms)) return "—";
  if (ms < 1000) return `${Math.round(ms)} ms`;
  return `${(ms / 1000).toFixed(1)} s`;
}

export function formatSuccessRate(rate: number | null | undefined): string {
  if (rate == null || Number.isNaN(rate)) return "—";
  return `${Math.round(rate * 100)}%`;
}

export function buildCreatePayloadFromTemplate(
  template: AutomationTemplate,
  timezone: string,
  timeOverride?: string,
): {
  name: string;
  description: string;
  template_id: string;
  timezone: string;
  trigger: { type: string; time: string };
  commands: AutomationCommand[];
  delivery: { channels: string[] };
  enabled: boolean;
} {
  return {
    name: template.name,
    description: template.description,
    template_id: template.id,
    timezone,
    trigger: {
      type: template.default_schedule.type || "daily",
      time: formatLocalTimeForInput(
        timeOverride ?? template.default_schedule.time ?? "09:00",
      ),
    },
    commands: template.commands.map((cmd) => ({
      command: cmd.command,
      variables: cmd.variables,
    })),
    delivery: { channels: ["chat"] },
    enabled: true,
  };
}
