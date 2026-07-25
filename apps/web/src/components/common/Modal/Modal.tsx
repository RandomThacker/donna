"use client";

import { useEffect, useId, useRef } from "react";
import { createPortal } from "react-dom";

import { cn } from "@/lib/cn";

import { modalStyles as styles } from "./Modal.styles";
import type { ModalProps } from "./Modal.types";

export function Modal({
  open,
  onClose,
  title,
  description,
  children,
  labelledBy,
  describedBy,
}: ModalProps) {
  const titleId = useId();
  const descriptionId = useId();
  const panelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      return;
    }

    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    window.addEventListener("keydown", onKeyDown);
    panelRef.current?.focus();

    return () => {
      document.body.style.overflow = previousOverflow;
      window.removeEventListener("keydown", onKeyDown);
    };
  }, [open, onClose]);

  if (!open || typeof document === "undefined") {
    return null;
  }

  return createPortal(
    <div className={styles.root} role="presentation">
      <button
        type="button"
        className={styles.backdrop}
        aria-label="Close dialog"
        onClick={onClose}
      />
      <div
        ref={panelRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby={labelledBy ?? titleId}
        aria-describedby={
          description ? (describedBy ?? descriptionId) : describedBy
        }
        tabIndex={-1}
        className={styles.panel}
      >
        <button
          type="button"
          className={styles.close}
          aria-label="Close"
          onClick={onClose}
        >
          <span aria-hidden>×</span>
        </button>
        <header className={styles.header}>
          <h2 id={labelledBy ?? titleId} className={styles.title}>
            {title}
          </h2>
          {description ? (
            <p
              id={describedBy ?? descriptionId}
              className={styles.description}
            >
              {description}
            </p>
          ) : null}
        </header>
        <div className={cn(styles.body)}>{children}</div>
      </div>
    </div>,
    document.body,
  );
}
