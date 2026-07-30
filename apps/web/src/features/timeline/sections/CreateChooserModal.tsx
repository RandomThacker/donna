"use client";

import { Modal } from "@/components/common";

type Props = {
  open: boolean;
  onClose: () => void;
  onEvent: () => void;
  onReminder: () => void;
};

export function CreateChooserModal({
  open,
  onClose,
  onEvent,
  onReminder,
}: Props) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Create"
      description="Add something to your timeline."
    >
      <div className="grid gap-2">
        <button
          type="button"
          className="rounded-xl border border-donna-border bg-donna-surface-2 px-4 py-3 text-left text-donna-text transition-colors hover:border-donna-accent/40 hover:bg-donna-accent-soft"
          onClick={onEvent}
        >
          <p className="font-medium">Event</p>
          <p className="mt-0.5 text-xs text-donna-muted">
            Meeting or block of time
          </p>
        </button>
        <button
          type="button"
          className="rounded-xl border border-donna-border bg-donna-surface-2 px-4 py-3 text-left text-donna-text transition-colors hover:border-donna-accent/40 hover:bg-donna-accent-soft"
          onClick={onReminder}
        >
          <p className="font-medium">Reminder</p>
          <p className="mt-0.5 text-xs text-donna-muted">Ping at a moment</p>
        </button>
      </div>
    </Modal>
  );
}
