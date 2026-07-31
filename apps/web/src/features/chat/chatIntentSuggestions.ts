import { commandGuides } from "@/features/commands";

export type ChatIntentSuggestion = {
  id: string;
  /** Short chip label (1–3 words). */
  label: string;
  /** Full commands filled into the composer on tap. */
  phrases: string[];
};

/** Short labels for the chat suggestion strip — keep in sync with commandGuides. */
const LABELS: Record<string, string> = {
  greeting: "Say hello",
  "create-task": "Create a task",
  "complete-task": "Complete a task",
  "create-reminder": "Create a reminder",
  "create-event": "Create an event",
  "query-today": "What's today",
  "query-tomorrow": "What's tomorrow",
  "query-due-today": "Due today",
};

export const chatIntentSuggestions: ChatIntentSuggestion[] = commandGuides
  .map((guide) => {
    const phrases = guide.examples.map((example) => example.phrase).filter(Boolean);
    if (phrases.length === 0) return null;
    return {
      id: guide.id,
      label: LABELS[guide.id] ?? guide.title,
      phrases,
    };
  })
  .filter((item): item is ChatIntentSuggestion => item !== null);

export function pickSuggestionPhrase(suggestion: ChatIntentSuggestion): string {
  const index = Math.floor(Math.random() * suggestion.phrases.length);
  return suggestion.phrases[index] ?? suggestion.phrases[0] ?? "";
}
