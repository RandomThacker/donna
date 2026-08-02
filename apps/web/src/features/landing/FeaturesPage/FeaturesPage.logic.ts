import type { FeaturesCopy } from "./FeaturesPage.types";

export const featuresCopy: FeaturesCopy = {
  eyebrow: "Features",
  headlineLead: "A personal assistant,",
  headlineEmphasis: "not another chatbot.",
  subhead:
    "Donna plans your morning, guards your calendar, and follows up on what you said you'd do — with a voice you choose and systems that never forget.",
  whoEyebrow: "Who she is",
  whoTitle: "The one who runs your day.",
  whoBody: [
    "Inspired by Donna Paulsen from Suits — the assistant who ran the day so Harvey could win the room. This Donna does the same for yours: proactive, calm, and never guilt-inducing.",
    "You message her like a person. She replies with the personality you pick, while the work underneath stays reliable, rule-based, and yours.",
  ],
  pillars: [
    {
      id: "one-surface",
      title: "One surface",
      body: "Dashboard, calendar, tasks, notes, and chat in a single place.",
    },
    {
      id: "reaches-first",
      title: "She reaches first",
      body: "Briefings and check-ins arrive before you think to ask.",
    },
    {
      id: "never-forgets",
      title: "Never forgets",
      body: "Commitments carry forward instead of quietly disappearing.",
    },
  ],
  highlightsEyebrow: "What sets her apart",
  highlightsTitle: "The parts you won't find elsewhere.",
  highlightsDescription:
    "Four things that make Donna feel less like software and more like someone on your side.",
  highlights: [
    {
      id: "personality",
      icon: "spark",
      kicker: "Personality engine",
      title: "She sounds like someone.",
      body: "Pick professional, casual, or flirty. Donna adapts nicknames, punchlines, and emoji level — while the underlying answer stays exactly as accurate.",
      points: [
        "Three personalities, tuned per user",
        "Tone changes, outcomes never do",
        "Greetings that match your time of day",
      ],
    },
    {
      id: "automations",
      icon: "repeat",
      kicker: "Proactive automations",
      title: "She starts the conversation.",
      body: "Schedule the morning brief, task review, or evening reflection. Donna runs them on your local clock and delivers to chat and your phone.",
      points: [
        "Daily or specific weekdays",
        "Chat plus web push delivery",
        "Full run history when you want the receipts",
      ],
    },
    {
      id: "command-chat",
      icon: "send",
      kicker: "Command chat",
      title: "Messaging that actually does things.",
      body: "An iMessage-style thread where every phrase maps to a real action. No invented meetings, no hallucinated tasks — just commands that land.",
      points: [
        "Create tasks, events, and reminders by text",
        "Ask what's today, tomorrow, or still open",
        "A visible palette of everything she understands",
      ],
    },
    {
      id: "insights",
      icon: "info",
      kicker: "Donna noticed",
      title: "Observations, not dashboards.",
      body: "She reads your real calendar and task list to surface the things worth knowing — open focus windows, today's progress, tomorrow's load.",
      points: [
        "Finds the gaps you can actually protect",
        "Flags a heavy tomorrow before tonight",
        "Quiet when there's genuinely nothing to say",
      ],
    },
  ],
  coreEyebrow: "The essentials",
  coreTitle: "Everything a day needs.",
  coreDescription:
    "The daily surfaces underneath — fast, quiet, and built to stay out of your way.",
  core: [
    {
      id: "dashboard",
      icon: "home",
      title: "Home dashboard",
      blurb:
        "Greeting, next meeting, today's timeline, and quick tasks at a glance.",
    },
    {
      id: "calendar",
      icon: "calendar",
      title: "Unified calendar",
      blurb:
        "Google and Microsoft in one honest agenda, plus events Donna owns.",
    },
    {
      id: "tasks",
      icon: "tasks",
      title: "Daily todo",
      blurb: "A focused journal for today, with backlog when things slip.",
    },
    {
      id: "reminders",
      icon: "bell",
      title: "Reminders",
      blurb:
        "One-shot or weekly, delivered by push on your local time — not the server's.",
    },
    {
      id: "notes",
      icon: "notes",
      title: "Notes",
      blurb: "Fast, ceremony-free capture beside the rest of your day.",
    },
    {
      id: "integrations",
      icon: "link",
      title: "Integrations",
      blurb:
        "Connect work and personal calendars without touching your login account.",
    },
  ],
  horizonLabel: "On the horizon",
  horizon: ["Memories", "Full AI conversation", "Voice", "Telegram & WhatsApp"],
  closeEyebrow: "Begin",
  closeTitle: "Meet her in the morning.",
  closeBody:
    "Sign in, connect a calendar if you want, and let Donna take the first pass at your day.",
  closeCtaLabel: "Start with Google",
  footnote: "Built to feel like a person. Backed by systems that don't forget.",
};

export function getFeaturesContent() {
  return { copy: featuresCopy };
}
