"use client";

import { useCallback, useEffect, useMemo, useState } from "react";

import {
  createAutomation,
  deleteAutomation,
  fetchAutomationTemplates,
  fetchAutomations,
  previewAutomation,
  runAutomation,
  updateAutomation,
} from "./Automations.api";
import {
  AUTOMATION_WEEKDAYS,
  buildCreatePayloadFromTemplate,
  commandLabel,
  commandsPreview,
  defaultBrowserTimezone,
  formatDelivery,
  formatDurationMs,
  formatLocalTimeForInput,
  formatRunAt,
  formatSchedule,
  formatStatusLabel,
  formatSuccessRate,
  weekdayFromToday,
  type AutomationWeekdayCode,
} from "./Automations.logic";
import { automationsStyles as styles } from "./Automations.styles";
import type {
  Automation,
  AutomationRunResult,
  AutomationTemplate,
} from "./Automations.types";

export function AutomationsPanel() {
  const [automations, setAutomations] = useState<Automation[]>([]);
  const [templates, setTemplates] = useState<AutomationTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [preview, setPreview] = useState<{
    name: string;
    result: AutomationRunResult;
  } | null>(null);
  const [runNotice, setRunNotice] = useState<{
    name: string;
    result: AutomationRunResult;
  } | null>(null);
  const [templateId, setTemplateId] = useState("morning_brief");
  const [localTime, setLocalTime] = useState("09:00");
  const [scheduleType, setScheduleType] = useState<"daily" | "weekly">("daily");
  const [selectedDays, setSelectedDays] = useState<AutomationWeekdayCode[]>([
    weekdayFromToday(),
  ]);
  const [customCommands, setCustomCommands] = useState("");
  const timezone = defaultBrowserTimezone();

  const selectedTemplate = useMemo(
    () => templates.find((t) => t.id === templateId) ?? null,
    [templates, templateId],
  );

  const load = useCallback(async (signal?: AbortSignal) => {
    setLoading(true);
    setError(null);
    try {
      const [autosData, templatesData] = await Promise.all([
        fetchAutomations(signal),
        fetchAutomationTemplates(signal),
      ]);
      setAutomations(autosData.automations ?? []);
      const tmpls = templatesData.templates ?? [];
      setTemplates(tmpls);
      if (tmpls.length > 0 && !tmpls.some((t) => t.id === templateId)) {
        setTemplateId(tmpls[0]!.id);
      }
      const match = tmpls.find((t) => t.id === (templateId || tmpls[0]?.id));
      if (match?.default_schedule?.time) {
        setLocalTime(formatLocalTimeForInput(match.default_schedule.time));
      }
    } catch (err) {
      if (signal?.aborted) return;
      setError(err instanceof Error ? err.message : "Couldn’t load automations");
    } finally {
      if (!signal?.aborted) {
        setLoading(false);
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- initial template id only
  }, []);

  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, [load]);

  useEffect(() => {
    if (selectedTemplate?.default_schedule?.time) {
      setLocalTime(formatLocalTimeForInput(selectedTemplate.default_schedule.time));
    }
  }, [selectedTemplate?.id, selectedTemplate?.default_schedule?.time]);

  async function handleAdd() {
    if (!selectedTemplate) return;
    if (scheduleType === "weekly" && selectedDays.length === 0) {
      setError("Pick at least one day for a weekly schedule.");
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const payload = buildCreatePayloadFromTemplate(selectedTemplate, timezone, {
        type: scheduleType,
        time: localTime,
        days: scheduleType === "weekly" ? selectedDays : undefined,
      });
      if (selectedTemplate.id === "custom") {
        const commands = customCommands
          .split("\n")
          .map((line) => line.trim())
          .filter(Boolean);
        if (commands.length === 0) {
          setError("Add at least one command for a custom automation.");
          setSaving(false);
          return;
        }
        payload.commands = commands.map((message) => ({
          command: "chat_message",
          variables: { message },
        }));
        payload.name = payload.name || "Custom Automation";
      }
      const created = await createAutomation(payload);
      setAutomations((prev) =>
        [...prev, created].sort((a, b) =>
          (a.trigger?.time ?? "").localeCompare(b.trigger?.time ?? ""),
        ),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t add automation");
    } finally {
      setSaving(false);
    }
  }

  function toggleDay(code: AutomationWeekdayCode) {
    setSelectedDays((prev) => {
      const set = new Set(prev);
      if (set.has(code)) {
        set.delete(code);
      } else {
        set.add(code);
      }
      return AUTOMATION_WEEKDAYS.map((d) => d.code).filter((c) =>
        set.has(c),
      ) as AutomationWeekdayCode[];
    });
  }

  async function handleToggle(auto: Automation) {
    setError(null);
    try {
      const updated = await updateAutomation(auto.id, { enabled: !auto.enabled });
      setAutomations((prev) =>
        prev.map((item) => (item.id === auto.id ? updated : item)),
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t update automation");
    }
  }

  async function handleDelete(auto: Automation) {
    setError(null);
    try {
      await deleteAutomation(auto.id);
      setAutomations((prev) => prev.filter((item) => item.id !== auto.id));
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t delete automation");
    }
  }

  async function handleRunNow(auto: Automation) {
    setError(null);
    setBusyId(auto.id);
    try {
      const result = await runAutomation(auto.id);
      setRunNotice({ name: auto.name, result });
      await load();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t run automation");
    } finally {
      setBusyId(null);
    }
  }

  async function handlePreview(auto: Automation) {
    setError(null);
    setBusyId(auto.id);
    try {
      const result = await previewAutomation(auto.id);
      setPreview({ name: auto.name, result });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Couldn’t preview automation");
    } finally {
      setBusyId(null);
    }
  }

  return (
    <>
      <div className={styles.list}>
        {loading ? (
          <p className={styles.empty}>Loading automations…</p>
        ) : automations.length === 0 ? (
          <p className={styles.empty}>
            No automations yet. Pick a template below — Donna will run those
            commands on a schedule and post into chat.
          </p>
        ) : (
          automations.map((auto) => (
            <div key={auto.id} className={styles.card}>
              <div className={styles.cardTop}>
                <div>
                  <p className={styles.title}>{auto.name}</p>
                  {auto.description ? (
                    <p className={styles.description}>{auto.description}</p>
                  ) : null}
                </div>
                <div className={styles.actions}>
                  <button
                    type="button"
                    className={styles.secondaryBtn}
                    disabled={busyId === auto.id}
                    onClick={() => void handleRunNow(auto)}
                  >
                    ▶ Run Now
                  </button>
                  <button
                    type="button"
                    className={styles.secondaryBtn}
                    disabled={busyId === auto.id}
                    onClick={() => void handlePreview(auto)}
                  >
                    👁 Preview
                  </button>
                  <button
                    type="button"
                    className={styles.toggle}
                    onClick={() => void handleToggle(auto)}
                  >
                    {auto.enabled ? "Enabled" : "Paused"}
                  </button>
                  <button
                    type="button"
                    className={styles.dangerBtn}
                    onClick={() => void handleDelete(auto)}
                  >
                    Delete
                  </button>
                </div>
              </div>
              <div className={styles.metaGrid}>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Schedule · </span>
                  {formatSchedule(auto)} · {auto.timezone}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Delivery · </span>
                  {formatDelivery(auto)}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Last run · </span>
                  {formatRunAt(auto.last_run_at)}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Status · </span>
                  {formatStatusLabel(auto.last_status)}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Duration · </span>
                  {formatDurationMs(auto.average_duration_ms)}
                  {auto.success_rate != null
                    ? ` · ${formatSuccessRate(auto.success_rate)} success`
                    : ""}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Commands · </span>
                  {auto.last_commands_total != null
                    ? `${auto.last_commands_success ?? 0}/${auto.last_commands_total}`
                    : commandsPreview(auto.commands ?? [])}
                </p>
                <p className={styles.metaItem}>
                  <span className={styles.metaLabel}>Next run · </span>
                  {formatRunAt(auto.next_run_at)}
                </p>
              </div>
            </div>
          ))
        )}
      </div>

      <div className={styles.form}>
        <div className={styles.formRow}>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="automation-template">
              Template
            </label>
            <select
              id="automation-template"
              className={styles.select}
              value={templateId}
              onChange={(event) => setTemplateId(event.target.value)}
            >
              {templates.map((tmpl) => (
                <option key={tmpl.id} value={tmpl.id}>
                  {tmpl.name}
                </option>
              ))}
            </select>
          </div>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="automation-schedule">
              Schedule
            </label>
            <select
              id="automation-schedule"
              className={styles.select}
              value={scheduleType}
              onChange={(event) => {
                const next = event.target.value === "weekly" ? "weekly" : "daily";
                setScheduleType(next);
                if (next === "weekly" && selectedDays.length === 0) {
                  setSelectedDays([weekdayFromToday()]);
                }
              }}
            >
              <option value="daily">Every day</option>
              <option value="weekly">Selected days</option>
            </select>
          </div>
          <div className={styles.field}>
            <label className={styles.label} htmlFor="automation-time">
              Time
            </label>
            <input
              id="automation-time"
              type="time"
              className={styles.timeInput}
              value={formatLocalTimeForInput(localTime)}
              onChange={(event) => setLocalTime(event.target.value)}
            />
          </div>
          <button
            type="button"
            className={styles.addBtn}
            disabled={saving || loading || !selectedTemplate}
            onClick={() => void handleAdd()}
          >
            {saving ? "Adding…" : "Add"}
          </button>
        </div>
        {scheduleType === "weekly" ? (
          <div className={styles.field}>
            <p className={styles.label}>Repeat on</p>
            <div className={styles.dayRow}>
              {AUTOMATION_WEEKDAYS.map((day) => {
                const active = selectedDays.includes(day.code);
                return (
                  <button
                    key={day.code}
                    type="button"
                    aria-pressed={active}
                    className={active ? styles.dayChipActive : styles.dayChip}
                    onClick={() => toggleDay(day.code)}
                  >
                    {day.label}
                  </button>
                );
              })}
            </div>
            {selectedDays.length === 0 ? (
              <p className={styles.scheduleHint}>Pick at least one day</p>
            ) : null}
          </div>
        ) : null}
        {selectedTemplate?.id === "custom" ? (
          <div className={styles.field}>
            <label className={styles.label} htmlFor="automation-commands">
              Commands (one per line)
            </label>
            <textarea
              id="automation-commands"
              className={styles.textInput}
              rows={3}
              placeholder={"What do I have today?\nWhat's due today?"}
              value={customCommands}
              onChange={(event) => setCustomCommands(event.target.value)}
            />
          </div>
        ) : selectedTemplate ? (
          <p className={styles.commandsHint}>
            {selectedTemplate.description} ·{" "}
            {commandsPreview(selectedTemplate.commands)}
          </p>
        ) : null}
      </div>

      {runNotice ? (
        <div className={styles.notice}>
          <p className={styles.noticeTitle}>
            Ran {runNotice.name} · {formatStatusLabel(runNotice.result.status)} ·{" "}
            {formatDurationMs(runNotice.result.duration_ms)}
          </p>
          <pre className={styles.noticeBody}>{runNotice.result.response}</pre>
          <button
            type="button"
            className={styles.secondaryBtn}
            onClick={() => setRunNotice(null)}
          >
            Dismiss
          </button>
        </div>
      ) : null}

      {error ? <p className={styles.error}>{error}</p> : null}
      <p className={styles.note}>
        Uses your device timezone ({timezone}). Donna runs each command through
        chat and posts one combined reply.{" "}
        <a href="/dashboard/automations" className="underline underline-offset-2">
          View history
        </a>
      </p>

      {preview ? (
        <div
          className={styles.modalBackdrop}
          role="presentation"
          onClick={() => setPreview(null)}
        >
          <div
            className={styles.modal}
            role="dialog"
            aria-modal="true"
            aria-label={`${preview.name} preview`}
            onClick={(event) => event.stopPropagation()}
          >
            <div className={styles.modalHeader}>
              <div>
                <p className={styles.modalEyebrow}>Preview</p>
                <h3 className={styles.modalTitle}>{preview.name}</h3>
              </div>
              <button
                type="button"
                className={styles.secondaryBtn}
                onClick={() => setPreview(null)}
              >
                Close
              </button>
            </div>
            <p className={styles.modalMeta}>
              {formatStatusLabel(preview.result.status)} ·{" "}
              {formatDurationMs(preview.result.duration_ms)} ·{" "}
              {preview.result.commands_success}/{preview.result.commands_total}{" "}
              commands · not saved to history
            </p>
            <pre className={styles.modalResponse}>{preview.result.response}</pre>
            <div className={styles.commandList}>
              {preview.result.commands.map((cmd) => (
                <div key={`${cmd.order_index}-${cmd.command}`} className={styles.commandRow}>
                  <div>
                    <p className={styles.commandName}>
                      {commandLabel({
                        command: cmd.command_key || cmd.command,
                        label: cmd.command,
                      })}
                    </p>
                    <p className={styles.commandMeta}>
                      {formatStatusLabel(cmd.status)} ·{" "}
                      {formatDurationMs(cmd.duration_ms)}
                      {cmd.error ? ` · ${cmd.error}` : ""}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
