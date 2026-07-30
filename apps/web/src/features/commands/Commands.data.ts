import type { IconName } from "@/components/common";

export type CommandExample = {
  phrase: string;
  note?: string;
};

export type CommandGuide = {
  id: string;
  intent: string;
  title: string;
  blurb: string;
  icon: IconName;
  examples: CommandExample[];
};

/** Canonical MVP phrases that RuleBasedParser accepts — keep in sync with docs/COMMAND_CHAT.md */
export const commandGuides: CommandGuide[] = [
  {
    id: "create-task",
    intent: "CREATE_TASK",
    title: "Create a task",
    blurb: "Drops a task on today’s journal.",
    icon: "tasks",
    examples: [
      { phrase: "Add task Finish Timeline UI" },
      { phrase: "Create task Ship notifications" },
      { phrase: "Todo call the dentist" },
    ],
  },
  {
    id: "complete-task",
    intent: "COMPLETE_TASK",
    title: "Complete a task",
    blurb: "Matches by title against today’s open tasks.",
    icon: "check",
    examples: [
      { phrase: "Complete task Finish Timeline UI" },
      { phrase: "Mark Finish API done" },
      { phrase: "Finish task Ship notifications" },
    ],
  },
  {
    id: "create-reminder",
    intent: "CREATE_REMINDER",
    title: "Create a reminder",
    blurb: "One-shot or weekly — Donna picks up time and recurrence.",
    icon: "clock",
    examples: [
      { phrase: "Remind me tomorrow at 6 PM to stretch" },
      { phrase: "Remind me every Friday at 8 PM to call Mom" },
      { phrase: "Create reminder to water plants today at 9 AM" },
    ],
  },
  {
    id: "create-event",
    intent: "CREATE_EVENT",
    title: "Create an event",
    blurb: "Adds a Donna calendar event (default length: one hour).",
    icon: "calendar",
    examples: [
      { phrase: "Schedule meeting Standup tomorrow at 10 AM" },
      { phrase: "Create event Guitar class today at 7 PM" },
      { phrase: "Meeting Design review tomorrow at 2 PM" },
    ],
  },
  {
    id: "query-today",
    intent: "QUERY_TODAY",
    title: "What’s on today",
    blurb: "Timeline for today — events and reminders.",
    icon: "sun",
    examples: [
      { phrase: "What do I have today?" },
      { phrase: "What's on today" },
      { phrase: "Show today" },
    ],
  },
  {
    id: "query-tomorrow",
    intent: "QUERY_TOMORROW",
    title: "What’s on tomorrow",
    blurb: "Same idea, next civil day.",
    icon: "sunrise",
    examples: [
      { phrase: "What do I have tomorrow?" },
      { phrase: "What's on tomorrow" },
      { phrase: "Show tomorrow" },
    ],
  },
  {
    id: "query-due-today",
    intent: "QUERY_DUE_TODAY",
    title: "What’s due today",
    blurb: "Open tasks on today’s journal — not the calendar.",
    icon: "spark",
    examples: [
      { phrase: "What's due today?" },
      { phrase: "Due today" },
      { phrase: "Show my tasks today" },
    ],
  },
];
