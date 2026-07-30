"use client";

import { cn } from "@/lib/cn";

import { Modal } from "../Modal";
import { confirmDialogStyles as styles } from "./ConfirmDialog.styles";
import type { ConfirmDialogProps } from "./ConfirmDialog.types";

export function ConfirmDialog({
  open,
  title,
  description,
  confirmLabel = "Confirm",
  cancelLabel = "Cancel",
  confirming = false,
  tone = "default",
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal
      open={open}
      onClose={onCancel}
      title={title}
      description={description}
      className={styles.root}
    >
      <div className={styles.actions}>
        <button
          type="button"
          className={styles.cancel}
          disabled={confirming}
          onClick={onCancel}
        >
          {cancelLabel}
        </button>
        <button
          type="button"
          className={cn(
            styles.confirm,
            tone === "danger" ? styles.confirmDanger : styles.confirmDefault,
          )}
          disabled={confirming}
          onClick={onConfirm}
        >
          {confirming ? "Working…" : confirmLabel}
        </button>
      </div>
    </Modal>
  );
}
