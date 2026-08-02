"use client";

import { useEffect, useMemo, useState } from "react";
import { format } from "date-fns";

import { Modal } from "@/components/common";

import type {
  CreateDonnaEventInput,
  TimelineItem,
  UpdateDonnaEventInput,
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
  onCreate: (body: CreateDonnaEventInput) => Promise<void>;
  onUpdate: (id: string, body: UpdateDonnaEventInput) => Promise<void>;
};

function toLocalInput(date: Date): { date: string; time: string } {
  return {
    date: format(date, "yyyy-MM-dd"),
    time: format(date, "HH:mm"),
  };
}

function combine(date: string, time: string): string {
  return new Date(`${date}T${time}:00`).toISOString();
}

export function EventFormModal({
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
      const start = toLocalInput(new Date(editing.start_at));
      const end = toLocalInput(new Date(editing.end_at));
      return {
        title: editing.title,
        description: editing.description ?? "",
        date: start.date,
        startTime: start.time,
        endTime: end.time,
        timezone: editing.timezone || timezone,
        recurrence: editing.recurrence_rule ?? "",
      };
    }
    const base = day ?? new Date();
    const start = toLocalInput(base);
    const endDate = new Date(base);
    if (start.time === "00:00") {
      endDate.setHours(10, 0, 0, 0);
      start.time = "09:00";
    } else {
      endDate.setHours(endDate.getHours() + 1);
    }
    const end = toLocalInput(endDate);
    return {
      title: "",
      description: "",
      date: start.date,
      startTime: start.time,
      endTime: end.time,
      timezone,
      recurrence: "",
    };
  }, [editing, day, timezone, open]);

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [date, setDate] = useState("");
  const [startTime, setStartTime] = useState("");
  const [endTime, setEndTime] = useState("");
  const [tz, setTz] = useState(timezone);
  const [recurrence, setRecurrence] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setTitle(initial.title);
    setDescription(initial.description);
    setDate(initial.date);
    setStartTime(initial.startTime);
    setEndTime(initial.endTime);
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
    const startAt = combine(date, startTime);
    const endAt = combine(date, endTime);
    if (new Date(endAt) <= new Date(startAt)) {
      setError("End must be after start");
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
          start_at: startAt,
          end_at: endAt,
          timezone: tz,
          recurrence_rule: rule,
        });
      } else {
        await onCreate({
          title: trimmed,
          description: description.trim() || null,
          start_at: startAt,
          end_at: endAt,
          timezone: tz,
          all_day: false,
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
      title={editing ? "Edit event" : "Create event"}
      description="Donna-owned event — shows on Calendar and notifies before start."
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
            required
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
        <div className="grid grid-cols-1 gap-3.5 sm:grid-cols-2 sm:gap-3">
          <label className="block min-w-0">
            <span className="mb-1 block text-xs font-medium text-donna-muted">
              Start
            </span>
            <input
              type="time"
              className={formFieldClass}
              value={startTime}
              onChange={(e) => setStartTime(e.target.value)}
            />
          </label>
          <label className="block min-w-0">
            <span className="mb-1 block text-xs font-medium text-donna-muted">
              End
            </span>
            <input
              type="time"
              className={formFieldClass}
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
            />
          </label>
        </div>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Timezone
          </span>
          <input
            className={formFieldClass}
            value={tz}
            onChange={(e) => setTz(e.target.value)}
            placeholder="Asia/Kolkata"
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
