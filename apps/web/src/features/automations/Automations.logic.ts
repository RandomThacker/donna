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

export const AUTOMATION_WEEKDAYS = [
  { code: "MO", label: "Mon" },
  { code: "TU", label: "Tue" },
  { code: "WE", label: "Wed" },
  { code: "TH", label: "Thu" },
  { code: "FR", label: "Fri" },
  { code: "SA", label: "Sat" },
  { code: "SU", label: "Sun" },
] as const;

export type AutomationWeekdayCode =
  (typeof AUTOMATION_WEEKDAYS)[number]["code"];

export function weekdayFromToday(): AutomationWeekdayCode {
  const map: AutomationWeekdayCode[] = [
    "SU",
    "MO",
    "TU",
    "WE",
    "TH",
    "FR",
    "SA",
  ];
  return map[new Date().getDay()] ?? "MO";
}

export function formatWeekdayCodes(days: string[] | undefined): string {
  if (!days || days.length === 0) return "";
  const order = AUTOMATION_WEEKDAYS.map((d) => d.code);
  const labels = [...days]
    .sort((a, b) => order.indexOf(a as AutomationWeekdayCode) - order.indexOf(b as AutomationWeekdayCode))
    .map((code) => AUTOMATION_WEEKDAYS.find((d) => d.code === code)?.label ?? code);
  return labels.join(", ");
}

export function formatSchedule(auto: Automation): string {
  const time = formatLocalTimeForInput(auto.trigger?.time ?? "09:00");
  const type = (auto.trigger?.type ?? "daily").toLowerCase();
  if (type === "weekly") {
    const days = formatWeekdayCodes(auto.trigger?.days);
    return days ? `${days} at ${time}` : `Weekly at ${time}`;
  }
  return `Every day at ${time}`;
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

export type AutomationScheduleInput = {
  type: "daily" | "weekly";
  time?: string;
  days?: string[];
};

export function buildCreatePayloadFromTemplate(
  template: AutomationTemplate,
  timezone: string,
  schedule?: AutomationScheduleInput,
): {
  name: string;
  description: string;
  template_id: string;
  timezone: string;
  trigger: { type: string; time: string; days?: string[] };
  commands: AutomationCommand[];
  delivery: { channels: string[] };
  enabled: boolean;
} {
  const type = schedule?.type ?? "daily";
  const trigger: { type: string; time: string; days?: string[] } = {
    type,
    time: formatLocalTimeForInput(
      schedule?.time ?? template.default_schedule.time ?? "09:00",
    ),
  };
  if (type === "weekly") {
    trigger.days =
      schedule?.days && schedule.days.length > 0
        ? schedule.days
        : [weekdayFromToday()];
  }
  return {
    name: template.name,
    description: template.description,
    template_id: template.id,
    timezone,
    trigger,
    commands: template.commands.map((cmd) => ({
      command: cmd.command,
      variables: cmd.variables,
    })),
    delivery: { channels: ["chat"] },
    enabled: true,
  };
}
