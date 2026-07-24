import type { DashboardData } from "./Dashboard.types";

export const dashboardData: DashboardData = {
  nav: [
    { id: "home", label: "Home", icon: "home", href: "/dashboard", active: true },
    { id: "calendar", label: "Calendar", icon: "calendar", href: "/dashboard" },
    { id: "tasks", label: "Tasks", icon: "tasks", href: "/dashboard" },
    { id: "notes", label: "Notes", icon: "notes", href: "/dashboard" },
    { id: "memories", label: "Memories", icon: "memory", href: "/dashboard" },
    { id: "settings", label: "Settings", icon: "settings", href: "/dashboard" },
  ],
  profileName: "Aryan Thacker",
  profileInitials: "AT",
  greeting: {
    salutation: "Good morning",
    name: "Aryan",
    summary: "2 meetings · 1 reminder · 5h focus time",
    nudge: "Great day to finish Calendar Integration.",
  },
  focus: {
    goal: "Ship Calendar Integration",
    progress: 0.62,
    timeRemaining: "~3h left in focus window",
    ctaLabel: "Continue",
    ctaHref: "#phone",
  },
  timeline: [
    { time: "09:30", title: "Team Standup", kind: "meeting" },
    { time: "11:00", title: "Deep Work", kind: "focus" },
    { time: "14:00", title: "Client Meeting", kind: "meeting" },
    { time: "17:00", title: "Guitar Class", kind: "personal" },
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
        preview: "Perfect. I'll help keep us on track today.",
        time: "9:41 AM",
        unread: 2,
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
