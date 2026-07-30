"use client";

import Link from "next/link";
import { usePathname, useSearchParams } from "next/navigation";

import { useAuth } from "@/features/auth";
import { getDashboardContent } from "@/features/dashboard/Dashboard.logic";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";
import { IMessageChat } from "@/features/dashboard/sections/DashboardPhone/IMessageChat";

import { chatStyles as styles } from "./Chat.styles";

export function Chat() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const { user } = useAuth();
  const navItems = navItemsForPath(pathname);
  const prefill = searchParams.get("prefill")?.trim() ?? "";
  const { data } = getDashboardContent();
  const donna =
    data.phone.conversations.find((item) => item.id === "donna") ??
    data.phone.conversations[0]!;

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = (() => {
    const parts = profileName.trim().split(/\s+/).filter(Boolean);
    if (parts.length === 0) return "D";
    if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
    return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
  })();

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={navItems}
          profileName={profileName}
          profileInitials={profileInitials}
          profileEmail={user?.email}
          profileAvatarUrl={user?.avatar_url}
        />
        <main className={styles.workspace}>
          <header className={styles.header}>
            <h1 className={styles.title}>Donna</h1>
            <p className={styles.subtitle}>
              Command chat — same iMessage thread as the phone. History is saved.{" "}
              <Link
                href="/dashboard/commands"
                className="text-donna-accent underline-offset-2 hover:underline"
              >
                See all commands
              </Link>
            </p>
          </header>
          <div className="mx-auto flex min-h-0 w-full max-w-md flex-1 flex-col overflow-hidden border-x border-donna-hairline sm:max-w-lg">
            <IMessageChat
              conversation={donna}
              live
              initialDraft={prefill}
              showBack={false}
              onBack={() => undefined}
            />
          </div>
        </main>
      </div>
    </div>
  );
}
