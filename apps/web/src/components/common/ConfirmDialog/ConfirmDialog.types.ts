export type ConfirmDialogProps = {
  open: boolean;
  title: string;
  description?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  confirming?: boolean;
  /** Destructive styling for irreversible actions (delete, etc.). */
  tone?: "default" | "danger";
  onConfirm: () => void;
  onCancel: () => void;
};
