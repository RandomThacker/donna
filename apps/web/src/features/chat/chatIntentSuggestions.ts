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

/** Pill order in the chat strip (anything else follows in guide order). */
const ORDER = [
  "greeting",
  "query-due-today",
  "query-today",
  "query-tomorrow",
] as const;

export const chatIntentSuggestions: ChatIntentSuggestion[] = (() => {
  const byId = new Map(
    commandGuides
      .map((guide) => {
        const phrases = guide.examples
          .map((example) => example.phrase)
          .filter(Boolean);
        if (phrases.length === 0) return null;
        return [
          guide.id,
          {
            id: guide.id,
            label: LABELS[guide.id] ?? guide.title,
            phrases,
          } satisfies ChatIntentSuggestion,
        ] as const;
      })
      .filter((entry): entry is readonly [string, ChatIntentSuggestion] => entry !== null),
  );

  const pinned = ORDER.map((id) => byId.get(id)).filter(
    (item): item is ChatIntentSuggestion => item !== undefined,
  );
  const pinnedIds = new Set<string>(ORDER);
  const rest = commandGuides
    .map((guide) => byId.get(guide.id))
    .filter(
      (item): item is ChatIntentSuggestion =>
        item !== undefined && !pinnedIds.has(item.id),
    );

  return [...pinned, ...rest];
})();

export function pickSuggestionPhrase(suggestion: ChatIntentSuggestion): string {
  const index = Math.floor(Math.random() * suggestion.phrases.length);
  return suggestion.phrases[index] ?? suggestion.phrases[0] ?? "";
}
