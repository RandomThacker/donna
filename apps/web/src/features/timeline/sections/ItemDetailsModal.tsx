"use client";

import { format } from "date-fns";

import { Icon, Modal } from "@/components/common";

import type { TimelineItem } from "../Timeline.types";
import {
  isDonnaEditable,
  notificationPolicyLabel,
  sourceLabel,
} from "../Timeline.utils";

type Props = {
  item: TimelineItem | null;
  onClose: () => void;
  onEdit: (item: TimelineItem) => void;
  onDelete: (item: TimelineItem) => void;
  deleting?: boolean;
};

export function ItemDetailsModal({
  item,
  onClose,
  onEdit,
  onDelete,
  deleting,
}: Props) {
  const open = Boolean(item);
  const editable = item ? isDonnaEditable(item) : false;

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={item?.title ?? "Details"}
      description={sourceLabel(item ?? ({ source: "DONNA", type: "EVENT" } as TimelineItem))}
    >
      {item ? (
        <div className="space-y-4">
          <dl className="space-y-3 text-sm">
            <Row label="When">
              {item.all_day
                ? format(new Date(item.start_at), "MMM d, yyyy") + " · All day"
                : `${format(new Date(item.start_at), "MMM d, yyyy · h:mm a")} – ${format(new Date(item.end_at), "h:mm a")}`}
            </Row>
            <Row label="Timezone">{item.timezone}</Row>
            <Row label="Source">{sourceLabel(item)}</Row>
            <Row label="Type">{item.type === "REMINDER" ? "Reminder" : "Event"}</Row>
            <Row label="Status">{item.status}</Row>
            {item.description ? (
              <Row label="Description">{item.description}</Row>
            ) : null}
            <Row label="Recurrence">
              {item.is_recurring || item.recurrence_rule
                ? item.recurrence_rule || "Repeats"
                : "None"}
            </Row>
            <Row label="Notification">{notificationPolicyLabel(item)}</Row>
            {item.read_only ? (
              <Row label="Access">Read-only (provider)</Row>
            ) : null}
          </dl>
          <div className="flex flex-wrap gap-2 border-t border-donna-hairline pt-4">
            {editable ? (
              <>
                <button
                  type="button"
                  className="inline-flex items-center gap-1.5 rounded-full border border-donna-border px-3 py-2 text-sm"
                  onClick={() => onEdit(item)}
                >
                  <Icon name="compose" className="h-3.5 w-3.5" />
                  Edit
                </button>
                <button
                  type="button"
                  className="inline-flex items-center gap-1.5 rounded-full border border-rose-500/40 px-3 py-2 text-sm text-rose-300 hover:bg-rose-500/10"
                  disabled={deleting}
                  onClick={() => {
                    if (
                      window.confirm(
                        `Delete “${item.title}”? This can’t be undone.`,
                      )
                    ) {
                      onDelete(item);
                    }
                  }}
                >
                  <Icon name="trash" className="h-3.5 w-3.5" />
                  {deleting ? "Deleting…" : "Delete"}
                </button>
              </>
            ) : null}
            <button
              type="button"
              className="ml-auto rounded-full px-3 py-2 text-sm text-donna-muted"
              onClick={onClose}
            >
              Close
            </button>
          </div>
        </div>
      ) : null}
    </Modal>
  );
}

function Row({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <dt className="text-xs font-medium uppercase tracking-wide text-donna-faint">
        {label}
      </dt>
      <dd className="mt-0.5 whitespace-pre-wrap text-donna-text">{children}</dd>
    </div>
  );
}
