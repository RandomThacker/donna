import type { DashboardData } from "./Dashboard.types";

export const dashboardData: DashboardData = {
  nav: [
    { id: "home", label: "Home", icon: "home", href: "/dashboard", active: true },
    { id: "calendar", label: "Calendar", icon: "calendar", href: "/dashboard/calendar" },
    { id: "tasks", label: "Todo", icon: "tasks", href: "/dashboard/tasks" },
    { id: "notes", label: "Notes", icon: "notes", href: "/dashboard/notes" },
    { id: "memories", label: "Memories", icon: "memory", href: "/dashboard/memories" },
    {
      id: "integrations",
      label: "Integrations",
      icon: "link",
      href: "/dashboard/integrations",
    },
    { id: "settings", label: "Settings", icon: "settings", href: "/dashboard/settings" },
  ],
  profileName: "Aryan Thacker",
  profileInitials: "AT",
  greeting: {
    salutation: "Good morning",
    name: "Aryan",
    emoji: "☀️",
    summary: "0 meetings · 0 tasks · 0 reminders",
    nudge: "",
  },
  focus: {
    goal: "Ship Calendar Integration",
    progress: 0.62,
    timeRemaining: "~3h left in focus window",
    ctaLabel: "Continue",
    ctaHref: "#phone",
  },
  timeline: [
    { id: "mock-1", time: "09:30", title: "Team Standup" },
    { id: "mock-2", time: "11:00", title: "Deep Work" },
    { id: "mock-3", time: "14:00", title: "Client Meeting" },
    { id: "mock-4", time: "17:00", title: "Guitar Class" },
  ],
  insights: [
    { id: "1", text: "Two-hour focus window after lunch." },
    { id: "2", text: "80% of this week's goals done." },
    { id: "3", text: "Tomorrow looks busy — keep tonight light." },
  ],
  tasks: [
    { id: "t1", label: "Wire Google Calendar adapter", done: false },
    { id: "t2", label: "Review yesterday's notes", done: true },
    { id: "t3", label: "Confirm Friday dinner", done: false },
    { id: "t4", label: "Send standup update", done: false },
  ],
  phone: {
    conversations: [
      {
        id: "donna",
        name: "Donna",
        preview: "Tell Donna what to do…",
        time: "",
        unread: 0,
        messages: [
          {
            id: "m1",
            role: "donna",
            text: "Good morning ☀️",
            time: "9:38 AM",
          },
          {
            id: "m2",
            role: "donna",
            text: "Yesterday we finished authentication. What's today's biggest goal?",
            time: "9:38 AM",
          },
          {
            id: "m3",
            role: "user",
            text: "Let's finish Calendar integration.",
            time: "9:40 AM",
          },
          {
            id: "m4",
            role: "donna",
            text: "Perfect. I'll help keep us on track today.",
            time: "9:41 AM",
          },
        ],
      },
    ],
  },
};

export function getDashboardContent() {
  return { data: dashboardData };
}
