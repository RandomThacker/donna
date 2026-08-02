"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";

import { Icon } from "@/components/common";
import { useTheme } from "@/components/theme";
import { useAuth } from "@/features/auth";
import { AutomationsPanel } from "@/features/automations";
import { PersonalityPanel } from "@/features/settings/Personality";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import { settingsStyles as styles } from "./Settings.styles";

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

export function Settings() {
  const router = useRouter();
  const pathname = usePathname();
  const { user, signOut } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const nav = navItemsForPath(pathname);
  const [avatarFailed, setAvatarFailed] = useState(false);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);
  const email = user?.email?.trim() || null;
  const nextTheme = theme === "dark" ? "light" : "dark";

  useEffect(() => {
    setAvatarFailed(false);
  }, [user?.avatar_url]);

  const showAvatar = Boolean(user?.avatar_url) && !avatarFailed;

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
          <div className={styles.workspaceInner}>
            <header className={styles.intro}>
              <h1 className={styles.pageTitle}>Settings</h1>
              <p className={styles.body}>
                Appearance, personality, automations, account, and a shortcut to
                chat commands. Calendar connections live under Integrations.
              </p>
            </header>

            <section className={styles.section} aria-labelledby="appearance-heading">
              <div>
                <h2 id="appearance-heading" className={styles.sectionTitle}>
                  Appearance
                </h2>
                <p className={styles.sectionHint}>How Donna looks on this device.</p>
              </div>
              <div className={styles.card}>
                <div className={styles.row}>
                  <div className={styles.rowMain}>
                    <p className={styles.rowLabel}>Theme</p>
                    <p className={styles.rowMeta}>
                      Currently {theme === "dark" ? "dark" : "light"}
                    </p>
                  </div>
                  <button
                    type="button"
                    className={styles.themeBtn}
                    onClick={toggleTheme}
                    aria-label={`Switch to ${nextTheme} theme`}
                  >
                    <Icon
                      name={theme === "dark" ? "sunrise" : "moon"}
                      className="h-4 w-4"
                    />
                    <span>
                      {theme === "dark" ? "Light mode" : "Dark mode"}
                    </span>
                  </button>
                </div>
              </div>
            </section>

            <section className={styles.section} aria-labelledby="personality-heading">
              <div>
                <h2 id="personality-heading" className={styles.sectionTitle}>
                  Personality
                </h2>
                <p className={styles.sectionHint}>
                  How Donna talks — greetings, reminders, and chat tone. Business
                  outcomes stay the same.
                </p>
              </div>
              <div className={styles.card}>
                <PersonalityPanel />
              </div>
            </section>

            <section className={styles.section} aria-labelledby="automations-heading">
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div>
                  <h2 id="automations-heading" className={styles.sectionTitle}>
                    Automations
                  </h2>
                  <p className={styles.sectionHint}>
                    Scheduled Donna commands — agenda, tasks, or a custom set —
                    posted into chat at a local time.
                  </p>
                </div>
                <a
                  href="/dashboard/automations"
                  className={styles.themeBtn}
                >
                  View runs
                </a>
              </div>
              <div className={styles.card}>
                <AutomationsPanel />
              </div>
            </section>

            <section className={styles.section} aria-labelledby="commands-heading">
              <div>
                <h2 id="commands-heading" className={styles.sectionTitle}>
                  Commands
                </h2>
                <p className={styles.sectionHint}>
                  Phrases Donna understands in chat right now.
                </p>
              </div>
              <div className={styles.card}>
                <div className={styles.row}>
                  <div className={styles.rowMain}>
                    <p className={styles.rowLabel}>Command guide</p>
                    <p className={styles.rowMeta}>
                      Browse every working phrase — copy or try in chat
                    </p>
                  </div>
                  <Link href="/dashboard/commands" className={styles.themeBtn}>
                    <Icon name="compose" className="h-4 w-4" />
                    <span>Open</span>
                  </Link>
                </div>
              </div>
            </section>

            <section className={styles.section} aria-labelledby="account-heading">
              <div>
                <h2 id="account-heading" className={styles.sectionTitle}>
                  Account
                </h2>
                <p className={styles.sectionHint}>Your Donna account.</p>
              </div>
              <div className={styles.card}>
                <div className={styles.profile}>
                  <span className={styles.avatar}>
                    {showAvatar ? (
                      <img
                        src={user!.avatar_url!}
                        alt=""
                        className={styles.avatarImage}
                        referrerPolicy="no-referrer"
                        onError={() => setAvatarFailed(true)}
                      />
                    ) : (
                      profileInitials
                    )}
                  </span>
                  <div className={styles.profileMeta}>
                    <p className={styles.profileName}>{profileName}</p>
                    {email ? (
                      <p className={styles.profileEmail}>{email}</p>
                    ) : null}
                  </div>
                </div>
                <div className={styles.row}>
                  <div className={styles.rowMain}>
                    <p className={styles.rowLabel}>Session</p>
                    <p className={styles.rowMeta}>
                      Sign out on this device
                    </p>
                  </div>
                  <button
                    type="button"
                    className={styles.signOut}
                    onClick={() => {
                      void (async () => {
                        await signOut();
                        router.replace("/");
                      })();
                    }}
                  >
                    Sign out
                  </button>
                </div>
              </div>
            </section>
          </div>
        </main>
      </div>
    </div>
  );
}
