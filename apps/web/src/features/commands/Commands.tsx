"use client";

import Link from "next/link";
import { useCallback, useState } from "react";
import { usePathname } from "next/navigation";

import { Icon } from "@/components/common";
import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";
import { cn } from "@/lib/cn";

import { commandGuides } from "./Commands.data";
import { commandsStyles as styles } from "./Commands.styles";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "D";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

function chatHrefFor(phrase: string): string {
  return `/dashboard/chat?prefill=${encodeURIComponent(phrase)}`;
}

function CopyTryRow({ phrase }: { phrase: string }) {
  const [copied, setCopied] = useState(false);

  const onCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(phrase);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      // Clipboard may be denied — ignore quietly.
    }
  }, [phrase]);

  return (
    <div className={styles.exampleRow}>
      <p className={styles.phrase}>{phrase}</p>
      <div className={styles.actions}>
        <button
          type="button"
          className={styles.iconBtn}
          aria-label={copied ? "Copied" : "Copy phrase"}
          title={copied ? "Copied" : "Copy"}
          onClick={() => void onCopy()}
        >
          <Icon name={copied ? "check" : "compose"} className="h-3.5 w-3.5" />
        </button>
        <Link
          href={chatHrefFor(phrase)}
          className={styles.iconBtn}
          aria-label="Try in chat"
          title="Try in chat"
        >
          <Icon name="send" className="h-3.5 w-3.5" />
        </Link>
      </div>
    </div>
  );
}

export function Commands() {
  const pathname = usePathname();
  const { user } = useAuth();
  const nav = navItemsForPath(pathname);

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
          <div className={styles.atmosphere} aria-hidden />
          <div className={styles.inner}>
            <header className={styles.hero}>
              <p className={styles.eyebrow}>
                <Icon name="spark" className="h-3.5 w-3.5" />
                Command palette
              </p>
              <h1 className={styles.title}>What Donna understands</h1>
              <p className={styles.subtitle}>
                These are the phrases that work today — no guessing. Copy one,
                or send it straight into chat.
              </p>
              <div className={styles.ctaRow}>
                <Link href="/dashboard/chat" className={styles.primaryCta}>
                  Open chat
                  <Icon name="arrow" className="h-3.5 w-3.5" />
                </Link>
                <span className={styles.secondaryHint}>
                  {commandGuides.length} commands · rule-based · no AI yet
                </span>
              </div>
            </header>

            <section className={styles.list} aria-label="Supported commands">
              {commandGuides.map((guide, index) => (
                <article
                  key={guide.id}
                  className={cn(styles.guide)}
                  style={{ animationDelay: `${index * 40}ms` }}
                >
                  <div className={styles.guideGlow} aria-hidden />
                  <div className={styles.guideHead}>
                    <span className={styles.iconWrap} aria-hidden>
                      <Icon name={guide.icon} className="h-4 w-4" />
                    </span>
                    <div className={styles.guideCopy}>
                      <span className={styles.intentPill}>{guide.intent}</span>
                      <h2 className={styles.guideTitle}>{guide.title}</h2>
                      <p className={styles.guideBlurb}>{guide.blurb}</p>
                    </div>
                  </div>
                  <div className={styles.examples}>
                    {guide.examples.map((example) => (
                      <CopyTryRow key={example.phrase} phrase={example.phrase} />
                    ))}
                  </div>
                </article>
              ))}
            </section>

            <footer className={styles.footer}>
              <p>
                <span className={styles.footerStrong}>Anything else</span> gets
                a help reply — we keep the set small so every command stays
                reliable. More land when these feel effortless.
              </p>
            </footer>
          </div>
        </main>
      </div>
    </div>
  );
}
