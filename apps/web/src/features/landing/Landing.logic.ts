import type { Metadata } from "next";

import type { LandingCopy } from "./Landing.types";

export const siteMetadata: Metadata = {
  title: "Donna — Personal AI Operating System",
  description:
    "Donna is a proactive personal assistant that plans your day, keeps commitments, and clears the noise.",
};

export const landingCopy: LandingCopy = {
  brand: "Donna",
  navLinks: [
    { label: "Features", href: "/features" },
    { label: "How she works", href: "/#rhythm" },
  ],
  getStarted: { label: "Get started", href: "#sign-in" },
  hero: {
    eyebrow: "Personal AI Operating System",
    headlineLead: "Your day, handled",
    headlineEmphasis: "before you even ask.",
    support:
      "Donna plans your mornings, guards your calendar, and remembers every commitment — a calm, proactive assistant working quietly in the background.",
    primaryCta: { label: "Start with Google", href: "#sign-in" },
    secondaryCta: { label: "See how she works", href: "#rhythm" },
    ctaNote: "Set up in under five minutes",
    chips: [
      { icon: "sunrise", label: "Morning briefing ready" },
      { icon: "spark", label: "Powered by Donna" },
      { icon: "calendar", label: "Calendar synced" },
    ],
    cards: [
      {
        icon: "check",
        title: "Today's focus",
        value: "3 / 5",
        sub: "tasks done",
        chart: "line",
      },
      {
        icon: "clock",
        title: "Time saved",
        value: "4h 20m",
        sub: "this week",
        chart: "bars",
      },
    ],
  },
  stats: [
    { value: "6:30 AM", label: "Your morning briefing, ready before you wake" },
    { value: "3×", label: "Gentle check-ins across the day" },
    { value: "1 place", label: "For your whole day — no app-switching" },
  ],
  promiseEyebrow: "The name",
  promiseTitle: "Not another chatbot.",
  promiseBody:
    "Harvey Specter had Donna — the assistant who ran his day so he could win the room. Inspired by that spirit from Suits, this Donna is an operating system for yours: a dashboard in the morning, a conversation that follows through, and a memory that never forgets.",
  capabilitiesEyebrow: "Capabilities",
  capabilitiesTitle: "What she carries for you",
  capabilitiesDescription:
    "Everything a great assistant handles — planning, time, and memory — held together in one calm surface.",
  capabilities: [
    {
      icon: "sunrise",
      title: "Daily planning",
      description:
        "Morning goals, midday check-ins, and evening reflection — momentum without the nagging.",
    },
    {
      icon: "calendar",
      title: "Calendar that listens",
      description:
        "Meetings, free time, and reminders across every account, unified into one honest view.",
    },
    {
      icon: "spark",
      title: "Memory that sticks",
      description:
        "Projects, people, preferences, and promises — recalled at the exact moment they matter.",
    },
    {
      icon: "check",
      title: "Quiet accountability",
      description:
        "She remembers what you set out to do and follows up gently, so nothing slips.",
    },
  ],
  rhythmEyebrow: "A day with Donna",
  rhythmTitle: "She moves with your day.",
  rhythmDescription:
    "From the first hello to the last reflection, Donna stays a step ahead — proactive, never pushy.",
  rhythm: [
    {
      icon: "sunrise",
      time: "Morning",
      title: "The briefing",
      description:
        "Your goal for the day, meetings ahead, and the one thing worth focusing on first.",
    },
    {
      icon: "clock",
      time: "Midday",
      title: "The check-in",
      description:
        "A light nudge on progress, a reshuffle if plans change, space cleared for deep work.",
    },
    {
      icon: "moon",
      time: "Evening",
      title: "The reflection",
      description:
        "What got done, what carries forward, and tomorrow — planned before you close the laptop.",
    },
  ],
  closeEyebrow: "Begin",
  closeTitle: "Open Donna tomorrow morning.",
  closeBody:
    "Start the day with one place that already knows what matters — and quietly takes care of the rest.",
  closeCta: { label: "Start with Google", href: "#sign-in" },
};

export function getLandingContent() {
  return { copy: landingCopy };
}
