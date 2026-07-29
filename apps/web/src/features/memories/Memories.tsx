"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useMemo } from "react";

import { Icon } from "@/components/common";
import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import {
  MEMORY_HEADLINE,
  MEMORY_SUBHEAD,
  MEMORY_TEASERS,
  statusLineForToday,
} from "./Memories.logic";
import { memoriesStyles as styles } from "./Memories.styles";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) {
    return "D";
  }
  if (parts.length === 1) {
    return parts[0]!.slice(0, 2).toUpperCase();
  }
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

export function Memories() {
  const pathname = usePathname();
  const { user } = useAuth();
  const nav = navItemsForPath(pathname);
  const statusLine = useMemo(() => statusLineForToday(), []);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={nav}
          profileName={profileName}
          profileInitials={profileInitials}
          profileEmail={user?.email}
          profileAvatarUrl={user?.avatar_url}
        />
        <main className={styles.workspace}>
          <div className={styles.stage}>
            <div className={styles.grid} aria-hidden />
            <div className={styles.glowA} aria-hidden />
            <div className={styles.glowB} aria-hidden />

            <div className={styles.panel}>
              <div className={styles.panelGlow} aria-hidden />
              <div className="relative">
                <span className={styles.badge}>
                  <span className={styles.badgeDot} aria-hidden />
                  Coming soon
                </span>
                <p className={styles.eyebrow}>Memories</p>
                <h1 className={styles.headline}>
                  <span className={styles.emphasis}>{MEMORY_HEADLINE}</span>
                </h1>
                <p className={styles.subhead}>{MEMORY_SUBHEAD}</p>
                <p className={styles.status}>
                  <Icon name="spark" className="h-4 w-4 shrink-0 text-donna-accent" />
                  {statusLine}
                </p>

                <div className={styles.cards}>
                  {MEMORY_TEASERS.map((memory, index) => (
                    <article
                      key={memory.id}
                      className={`${styles.card} ${memory.tilt} animate-donna-fade-up`}
                      style={{ animationDelay: `${120 + index * 70}ms` }}
                    >
                      <p className={styles.cardTag}>{memory.tag}</p>
                      <h2 className={styles.cardTitle}>{memory.title}</h2>
                      <p className={styles.cardSnippet}>{memory.snippet}</p>
                    </article>
                  ))}
                </div>

                <div className={styles.actions}>
                  <Link href="/dashboard/notes" className={styles.primaryBtn}>
                    Fine, I&apos;ll use Notes
                  </Link>
                  <Link href="/dashboard" className={styles.secondaryBtn}>
                    Back to home
                  </Link>
                </div>
                <p className={styles.footnote}>
                  No memories were harmed in the making of this placeholder.
                  Yet.
                </p>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
