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

/** Canonical MVP phrases — keep in sync with docs/COMMAND_CHAT.md and automation templates. */
export const commandGuides: CommandGuide[] = [
  {
    id: "greeting",
    intent: "GREETING",
    title: "Say hello",
    blurb: "Gets a warm time-of-day hello plus a personality punchline.",
    icon: "spark",
    examples: [
      { phrase: "Hi" },
      { phrase: "Hello" },
      { phrase: "Hey Donna" },
    ],
  },
  {
    id: "morning-greeting",
    intent: "MORNING_GREETING",
    title: "Morning greeting",
    blurb: "Good morning with a punchline — same as the Morning Greeting automation.",
    icon: "sun",
    examples: [
      { phrase: "Good morning" },
      { phrase: "Morning greeting" },
    ],
  },
  {
    id: "evening-greeting",
    intent: "EVENING_GREETING",
    title: "Evening greeting",
    blurb: "Evening hello that asks how the day went.",
    icon: "sunrise",
    examples: [
      { phrase: "Good evening" },
      { phrase: "Evening greeting" },
      { phrase: "How was my day" },
    ],
  },
  {
    id: "goodnight-greeting",
    intent: "GOOD_NIGHT_GREETING",
    title: "Good night",
    blurb: "Wishes good night with a soft closing line.",
    icon: "spark",
    examples: [
      { phrase: "Good night" },
      { phrase: "Night greeting" },
    ],
  },
  {
    id: "morning-brief",
    intent: "MORNING_BRIEF",
    title: "Morning Brief",
    blurb: "Same stack as the Morning Brief automation — greeting, today's agenda, tasks due. Try each line.",
    icon: "sun",
    examples: [
      { phrase: "Good morning", note: "Greeting" },
      { phrase: "What do I have today?", note: "Agenda" },
      { phrase: "What's due today?", note: "Tasks" },
    ],
  },
  {
    id: "query-today",
    intent: "QUERY_TODAY",
    title: "Today's Agenda",
    blurb: "Timeline for today — events and reminders (Today's Agenda automation).",
    icon: "calendar",
    examples: [
      { phrase: "What do I have today?" },
      { phrase: "What's on today" },
      { phrase: "Show today" },
    ],
  },
  {
    id: "query-due-today",
    intent: "QUERY_DUE_TODAY",
    title: "Task Review",
    blurb: "Open tasks on today’s journal — Task Review automation.",
    icon: "tasks",
    examples: [
      { phrase: "What's due today?" },
      { phrase: "Due today" },
      { phrase: "Show my tasks today" },
    ],
  },
  {
    id: "query-tomorrow",
    intent: "QUERY_TOMORROW",
    title: "Tomorrow Prep",
    blurb: "Peek at tomorrow — Tomorrow Prep automation.",
    icon: "sunrise",
    examples: [
      { phrase: "What do I have tomorrow?" },
      { phrase: "What's on tomorrow" },
      { phrase: "Show tomorrow" },
    ],
  },
  {
    id: "evening-review",
    intent: "EVENING_REVIEW",
    title: "Evening Review",
    blurb: "Evening greeting, tasks due, then tomorrow — same as Evening Review automation.",
    icon: "clock",
    examples: [
      { phrase: "Evening greeting", note: "How was today?" },
      { phrase: "What's due today?", note: "Open tasks" },
      { phrase: "What do I have tomorrow?", note: "Tomorrow" },
    ],
  },
  {
    id: "upcoming-meetings",
    intent: "QUERY_TODAY",
    title: "Upcoming Meetings",
    blurb: "Today's calendar in one place — Upcoming Meetings automation.",
    icon: "calendar",
    examples: [
      { phrase: "What do I have today?" },
      { phrase: "Show today" },
    ],
  },
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
];
