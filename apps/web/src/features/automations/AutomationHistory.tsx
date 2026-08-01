"use client";

import { useCallback, useEffect, useState } from "react";
import { usePathname } from "next/navigation";

import { useAuth } from "@/features/auth";
import { navItemsForPath } from "@/features/dashboard/dashboardNav";
import { DashboardSidebar } from "@/features/dashboard/sections/DashboardSidebar";

import {
  fetchAutomationAnalytics,
  fetchAutomationExecution,
  fetchAutomationHistoryAll,
} from "./Automations.api";
import {
  formatDurationMs,
  formatRunAt,
  formatStatusLabel,
  formatSuccessRate,
} from "./Automations.logic";
import { automationHistoryStyles as styles } from "./AutomationHistory.styles";
import type {
  AutomationAnalytics,
  AutomationExecution,
} from "./Automations.types";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean);
  if (parts.length === 0) return "D";
  if (parts.length === 1) return parts[0]!.slice(0, 2).toUpperCase();
  return `${parts[0]![0] ?? ""}${parts[1]![0] ?? ""}`.toUpperCase();
}

export function AutomationHistory() {
  const pathname = usePathname();
  const { user } = useAuth();
  const nav = navItemsForPath(pathname);
  const [executions, setExecutions] = useState<AutomationExecution[]>([]);
  const [analytics, setAnalytics] = useState<AutomationAnalytics | null>(null);
  const [expanded, setExpanded] = useState<string | null>(null);
  const [details, setDetails] = useState<Record<string, AutomationExecution>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const profileName =
    user?.display_name?.trim() || user?.email?.split("@")[0] || "You";
  const profileInitials = initialsFrom(profileName);

  const load = useCallback(async (signal?: AbortSignal) => {
    setLoading(true);
    setError(null);
    try {
      const [history, stats] = await Promise.all([
        fetchAutomationHistoryAll(signal, 50),
        fetchAutomationAnalytics(signal),
      ]);
      setExecutions(history.executions ?? []);
      setAnalytics(stats);
    } catch (err) {
      if (signal?.aborted) return;
      setError(err instanceof Error ? err.message : "Couldn’t load history");
    } finally {
      if (!signal?.aborted) setLoading(false);
    }
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);

  async function toggleExpand(executionId: string) {
    if (expanded === executionId) {
      setExpanded(null);
      return;
    }
    setExpanded(executionId);
    if (details[executionId]) return;
    try {
      const detail = await fetchAutomationExecution(executionId);
      setDetails((prev) => ({ ...prev, [executionId]: detail }));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t load execution");
    }
  }

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
            <header>
              <h1 className={styles.title}>Automation history</h1>
              <p className={styles.subtitle}>
                Every scheduled run Donna completed — timing, command results,
                and delivery.
              </p>
            </header>

            {analytics ? (
              <div className={styles.stats}>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Total executions</span>
                  <span className={styles.statValue}>
                    {analytics.total_executions}
                  </span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Success rate</span>
                  <span className={styles.statValue}>
                    {formatSuccessRate(analytics.success_rate)}
                  </span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Avg duration</span>
                  <span className={styles.statValue}>
                    {formatDurationMs(analytics.average_duration_ms)}
                  </span>
                </div>
                <div className={styles.stat}>
                  <span className={styles.statLabel}>Most run</span>
                  <span className={styles.statValue}>
                    {analytics.most_frequent_automation_name ?? "—"}
                  </span>
                </div>
              </div>
            ) : null}

            {error ? <p className={styles.error}>{error}</p> : null}

            <div className={styles.list}>
              {loading ? (
                <p className={styles.empty}>Loading history…</p>
              ) : executions.length === 0 ? (
                <p className={styles.empty}>
                  No executions yet. Once an automation fires, it will show up
                  here.
                </p>
              ) : (
                executions.map((exec) => {
                  const detail = details[exec.id] ?? exec;
                  const isOpen = expanded === exec.id;
                  return (
                    <div key={exec.id} className={styles.card}>
                      <div className={styles.cardHeader}>
                        <div>
                          <p className={styles.cardTitle}>
                            {exec.automation_name ?? "Automation"}
                          </p>
                          <p className={styles.cardMeta}>
                            {formatRunAt(exec.started_at)} ·{" "}
                            {formatDurationMs(exec.duration_ms)} ·{" "}
                            {exec.commands_success}/{exec.commands_total} commands
                            · {formatStatusLabel(exec.status)}
                          </p>
                        </div>
                        <button
                          type="button"
                          className={styles.expandBtn}
                          onClick={() => void toggleExpand(exec.id)}
                        >
                          {isOpen ? "Hide" : "Commands"}
                        </button>
                      </div>
                      {isOpen ? (
                        <div className={styles.commands}>
                          {(detail.commands ?? []).length === 0 ? (
                            <p className={styles.cardMeta}>
                              Loading command results…
                            </p>
                          ) : (
                            detail.commands!.map((cmd) => (
                              <div key={cmd.id} className={styles.commandRow}>
                                <p className={styles.commandText}>{cmd.command}</p>
                                <p className={styles.commandStatus}>
                                  {formatStatusLabel(cmd.status)}
                                </p>
                                <p className={styles.commandStatus}>
                                  {formatDurationMs(cmd.duration_ms)}
                                  {cmd.error ? ` · ${cmd.error}` : ""}
                                </p>
                              </div>
                            ))
                          )}
                          {detail.debug ? (
                            <pre className={styles.debug}>
                              {JSON.stringify(detail.debug, null, 2)}
                            </pre>
                          ) : null}
                        </div>
                      ) : null}
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
