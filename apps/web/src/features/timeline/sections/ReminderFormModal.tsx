"use client";

import { useEffect, useMemo, useState } from "react";
import { format } from "date-fns";

import { Modal } from "@/components/common";

import type {
  CreateDonnaReminderInput,
  TimelineItem,
  UpdateDonnaReminderInput,
} from "../Timeline.types";
import { entityIdForMutation } from "../Timeline.utils";
import {
  formFieldClass,
  parseRecurrenceRule,
  RecurrenceField,
} from "./RecurrenceField";

type Props = {
  open: boolean;
  onClose: () => void;
  day: Date | null;
  editing?: TimelineItem | null;
  timezone: string;
  saving?: boolean;
  onCreate: (body: CreateDonnaReminderInput) => Promise<void>;
  onUpdate: (id: string, body: UpdateDonnaReminderInput) => Promise<void>;
};

function toLocalInput(date: Date): { date: string; time: string } {
  return {
    date: format(date, "yyyy-MM-dd"),
    time: format(date, "HH:mm"),
  };
}

export function ReminderFormModal({
  open,
  onClose,
  day,
  editing,
  timezone,
  saving,
  onCreate,
  onUpdate,
}: Props) {
  const initial = useMemo(() => {
    if (editing) {
      const at = toLocalInput(new Date(editing.start_at));
      return {
        title: editing.title,
        description: editing.description ?? "",
        date: at.date,
        time: at.time,
        timezone: editing.timezone || timezone,
        recurrence: editing.recurrence_rule ?? "",
      };
    }
    const base = day ?? new Date();
    const at = toLocalInput(base);
    return {
      title: "",
      description: "",
      date: at.date,
      time: at.time === "00:00" ? "09:00" : at.time,
      timezone,
      recurrence: "",
    };
  }, [editing, day, timezone, open]);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [time, setTime] = useState("");
  const [tz, setTz] = useState(timezone);
  const [recurrence, setRecurrence] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setTitle(initial.title);
    setDescription(initial.description);
    setDate(initial.date);
    setTime(initial.time);
    setTz(initial.timezone);
    setRecurrence(initial.recurrence);
    setError(null);
  }, [open, initial]);

  async function submit() {
    const trimmed = title.trim();
    if (!trimmed) {
      setError("Title is required");
      return;
    }
    const parsed = parseRecurrenceRule(recurrence);
    if (
      parsed.selectValue === "__custom__" &&
      parsed.customDays.length === 0
    ) {
      setError("Pick at least one day for custom recurrence");
      return;
    }
    const triggerAt = new Date(`${date}T${time}:00`).toISOString();
    const rule = recurrence.trim() ? recurrence.trim() : null;
    try {
      if (editing) {
        const id = entityIdForMutation(editing);
        if (!id) {
          setError("This item can't be edited");
          return;
        }
        await onUpdate(id, {
          title: trimmed,
          description: description.trim() || null,
          trigger_at: triggerAt,
          timezone: tz,
          recurrence_rule: rule,
        });
      } else {
        await onCreate({
          title: trimmed,
          description: description.trim() || null,
          trigger_at: triggerAt,
          timezone: tz,
          recurrence_rule: rule,
        });
      }
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={editing ? "Edit reminder" : "Create reminder"}
      description="Donna reminder — fires at the time you set."
    >
      <div className="space-y-3.5">
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Title
          </span>
          <input
            className={formFieldClass}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
          />
        </label>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Description
          </span>
          <textarea
            className={formFieldClass}
            rows={3}
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </label>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Date
          </span>
          <input
            type="date"
            className={formFieldClass}
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </label>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Time
          </span>
          <input
            type="time"
            className={formFieldClass}
            value={time}
            onChange={(e) => setTime(e.target.value)}
          />
        </label>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Timezone
          </span>
          <input
            className={formFieldClass}
            value={tz}
            onChange={(e) => setTz(e.target.value)}
          />
        </label>
        <RecurrenceField value={recurrence} onChange={setRecurrence} />
        {error ? <p className="text-sm text-rose-400">{error}</p> : null}
        <div className="sticky bottom-0 -mx-1 flex justify-end gap-2 border-t border-donna-hairline bg-donna-surface px-1 pt-3 pb-1">
          <button
            type="button"
            className="min-h-11 rounded-full px-4 py-2.5 text-sm text-donna-muted hover:text-donna-text sm:min-h-0 sm:py-2"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={saving}
            className="min-h-11 rounded-full bg-donna-accent px-5 py-2.5 text-sm font-medium text-donna-on-accent disabled:opacity-50 sm:min-h-0 sm:py-2"
            onClick={() => void submit()}
          >
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </Modal>
  );
}
