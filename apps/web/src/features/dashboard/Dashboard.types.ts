import type { IconName } from "@/components/common";

export type DashboardNavItem = {
  id: string;
  label: string;
  icon: IconName;
  href: string;
  active?: boolean;
};

export type DashboardGreeting = {
  salutation: string;
  name: string;
  summary: string;
  nudge: string;
};

export type DashboardFocus = {
  goal: string;
  progress: number;
  timeRemaining: string;
  ctaLabel: string;
  ctaHref: string;
};

export type DashboardTimelineItem = {
  time: string;
  title: string;
  kind: "meeting" | "focus" | "personal";
};

export type DashboardInsight = {
  id: string;
  text: string;
};

export type DashboardTask = {
  id: string;
  label: string;
  done: boolean;
};

export type IMessageBubble = {
  id: string;
  role: "donna" | "user";
  text: string;
  time?: string;
};

export type IMessageConversation = {
  id: string;
  name: string;
  preview: string;
  time: string;
  unread: number;
  messages: IMessageBubble[];
};

export type DashboardPhone = {
  conversations: IMessageConversation[];
};

export type DashboardData = {
  nav: DashboardNavItem[];
  profileName: string;
  profileInitials: string;
  greeting: DashboardGreeting;
  focus: DashboardFocus;
  timeline: DashboardTimelineItem[];
  insights: DashboardInsight[];
  tasks: DashboardTask[];
  phone: DashboardPhone;
};
