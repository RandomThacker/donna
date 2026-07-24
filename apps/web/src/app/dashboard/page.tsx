import type { Metadata } from "next";

import { Dashboard } from "@/features/dashboard";

export const metadata: Metadata = {
  title: "Dashboard — Donna",
  description: "Your day with Donna — focus, timeline, and a conversation that stays open.",
};

export default function DashboardPage() {
  return <Dashboard />;
}
