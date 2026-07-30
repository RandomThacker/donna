"use client";

import { useEffect, useRef, useState } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { TaskTag } from "../../Tasks.types";
import { tagPickerStyles as styles } from "./TaskTagPicker.styles";

type TaskTagPickerProps = {
  tags: TaskTag[];
  selectedIds: string[];
  disabled?: boolean;
  onChange: (tagIds: string[]) => void;
  /** Controlled open state (e.g. from row overflow menu). */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  /** Hide the + trigger when opened from another control. */
  hideTrigger?: boolean;
};

export function TaskTagPicker({
  tags,
  selectedIds,
  disabled,
  onChange,
  open: openProp,
  onOpenChange,
  hideTrigger,
}: TaskTagPickerProps) {
  const [uncontrolledOpen, setUncontrolledOpen] = useState(false);
  const controlled = openProp !== undefined;
  const open = controlled ? openProp : uncontrolledOpen;
  const setOpen = (next: boolean) => {
    if (!controlled) {
      setUncontrolledOpen(next);
    }
    onOpenChange?.(next);
  };
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    const onPointerDown = (event: MouseEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", onPointerDown);
    return () => document.removeEventListener("mousedown", onPointerDown);
  }, [open]);

  const toggleTag = (tagId: string) => {
    const next = selectedIds.includes(tagId)
      ? selectedIds.filter((id) => id !== tagId)
      : [...selectedIds, tagId];
    onChange(next);
    setOpen(false);
  };

  return (
    <div ref={rootRef} className={styles.root}>
      {!hideTrigger ? (
        <button
          type="button"
          className={styles.trigger}
          aria-label="Edit tags"
          disabled={disabled || tags.length === 0}
          onClick={() => setOpen(!open)}
        >
          <Icon name="plus" className="h-3.5 w-3.5" />
        </button>
      ) : null}
      {open ? (
        <div className={styles.menu} role="menu">
          {tags.length === 0 ? (
            <p className={styles.empty}>Create a tag first</p>
          ) : (
            <ul className={styles.list}>
              {tags.map((tag) => {
                const checked = selectedIds.includes(tag.id);
                return (
                  <li key={tag.id}>
                    <button
                      type="button"
                      role="menuitemcheckbox"
                      aria-checked={checked}
                      className={cn(styles.option, checked && styles.optionOn)}
                      onClick={() => toggleTag(tag.id)}
                    >
                      <span
                        className={styles.swatch}
                        style={{ backgroundColor: tag.color }}
                        aria-hidden
                      />
                      <span className={styles.optionName}>{tag.name}</span>
                      {checked ? (
                        <Icon name="check" className="h-3.5 w-3.5 shrink-0" />
                      ) : null}
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      ) : null}
    </div>
  );
}
