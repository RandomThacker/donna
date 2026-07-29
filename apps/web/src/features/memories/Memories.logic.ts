export type MemoryTeaser = {
  id: string;
  title: string;
  snippet: string;
  tag: string;
  tilt: string;
};

export const MEMORY_HEADLINE = "Your brain outsourced this part.";

export const MEMORY_SUBHEAD =
  "Memories is where Donna will stash the stuff you swear you'll remember — then gently remind you that you absolutely will not.";

export const MEMORY_STATUS_LINES = [
  "Currently teaching neurons to sync with the cloud.",
  "Indexing every time you said \"I'll remember that.\"",
  "Building a vault for brilliant 2 a.m. ideas and suspicious passwords.",
  "Donna's filing cabinet is almost ready. The drawers just need attitude.",
] as const;

export const MEMORY_TEASERS: MemoryTeaser[] = [
  {
    id: "milk",
    title: "Buy milk",
    snippet: "Urgent. Probably. It's been a while.",
    tag: "Expired · 847 days",
    tilt: "-rotate-2",
  },
  {
    id: "idea",
    title: "Genius startup idea",
    snippet: "Something with AI and… sandwiches?",
    tag: "2:14 AM thought",
    tilt: "rotate-1",
  },
  {
    id: "password",
    title: "Definitely secure password",
    snippet: "Not password123. Definitely not.",
    tag: "Top secret-ish",
    tilt: "rotate-2",
  },
  {
    id: "mom",
    title: "Call mom back",
    snippet: "She left a voicemail. You listened. You did not call.",
    tag: "Carried × 12",
    tilt: "-rotate-1",
  },
  {
    id: "dentist",
    title: "Dentist — rescheduled again",
    snippet: "Your teeth are fine. Your calendar is not.",
    tag: "Avoidance streak: 3",
    tilt: "rotate-3",
  },
];

export function statusLineForToday(date = new Date()): string {
  const index = date.getDate() % MEMORY_STATUS_LINES.length;
  return MEMORY_STATUS_LINES[index]!;
}
