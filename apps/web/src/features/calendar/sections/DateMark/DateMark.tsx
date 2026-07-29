"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { format } from "../../Calendar.utils";
import { dateMarkStyles as styles } from "./DateMark.styles";

type DateMarkProps = {
  date: Date;
  onClick?: () => void;
  className?: string;
  "aria-label"?: string;
};

export function DateMark({
  date,
  onClick,
  className,
  "aria-label": ariaLabel,
}: DateMarkProps) {
  const label =
    ariaLabel ?? format(date, "EEEE, MMMM d, yyyy");
  const content = (
    <>
      <span className={styles.icon} aria-hidden>
        <Icon name="calendar" className="h-4 w-4" />
      </span>
      <span className={styles.date}>
        <span className={styles.dayNumber}>{format(date, "d")}</span>
        <span className={styles.meta}>{format(date, "EEE · MMM")}</span>
      </span>
    </>
  );

  if (onClick) {
    return (
      <button
        type="button"
        className={cn(styles.root, className)}
        onClick={onClick}
        aria-label={label}
      >
        {content}
      </button>
    );
  }

  return (
    <div className={cn(styles.root, className)} aria-label={label}>
      {content}
    </div>
  );
}
