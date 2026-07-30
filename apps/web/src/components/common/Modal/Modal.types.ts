export type ModalProps = {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
  labelledBy?: string;
  describedBy?: string;
  /** Extra classes on the fixed root (e.g. higher z-index when stacked). */
  className?: string;
};
