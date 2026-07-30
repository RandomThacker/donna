"use client";

import { useEffect, useState } from "react";

import { Modal } from "@/components/common";

import type { TaskOccurrence } from "../Tasks.types";

const fieldClass =
  "w-full rounded-xl border border-donna-border bg-donna-surface-2 px-3 py-2 text-sm text-donna-text outline-none transition-colors focus:border-donna-accent/50";

type Props = {
  open: boolean;
  occurrence: TaskOccurrence | null;
  saving?: boolean;
  onClose: () => void;
  onSave: (input: { title: string; date: string }) => Promise<void>;
};

export function TaskEditModal({
  open,
  occurrence,
  saving,
  onClose,
  onSave,
}: Props) {
  const [title, setTitle] = useState("");
  const [date, setDate] = useState("");
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open || !occurrence) return;
    setTitle(occurrence.title);
    setDate(occurrence.date);
    setError(null);
  }, [open, occurrence]);

  async function submit() {
    const trimmed = title.trim();
    if (!trimmed) {
      setError("Title is required");
      return;
    }
    if (!date) {
      setError("Date is required");
      return;
    }
    try {
      await onSave({ title: trimmed, date });
      onClose();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Save failed");
    }
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Edit task"
      description="Update the title or move it to another day."
    >
      <div className="space-y-3">
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Title
          </span>
          <input
            className={fieldClass}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
          />
        </label>
        <label className="block">
          <span className="mb-1 block text-xs font-medium text-donna-muted">
            Date
          </span>
          <input
            type="date"
            className={fieldClass}
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
        </label>
        {error ? <p className="text-sm text-rose-400">{error}</p> : null}
        <div className="flex justify-end gap-2 pt-2">
          <button
            type="button"
            className="rounded-full px-4 py-2 text-sm text-donna-muted hover:text-donna-text"
            onClick={onClose}
          >
            Cancel
          </button>
          <button
            type="button"
            disabled={saving}
            className="rounded-full bg-donna-accent px-4 py-2 text-sm font-medium text-donna-on-accent disabled:opacity-50"
            onClick={() => void submit()}
          >
            {saving ? "Saving…" : "Save"}
          </button>
        </div>
      </div>
    </Modal>
  );
}
