"use client";

import { useState } from "react";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { TAG_COLOR_PRESETS } from "../../Tasks.tags.types";
import type { TaskTag } from "../../Tasks.types";
import { TaskTagPill } from "../TaskTagPill";
import { tagsPanelStyles as styles } from "./TaskTagsPanel.styles";

type TaskTagsPanelProps = {
  tags: TaskTag[];
  filterTagIds: string[];
  isSaving?: boolean;
  onToggleFilter: (tagId: string) => void;
  onClearFilters: () => void;
  onCreateTag: (input: { name: string; color: string }) => void;
  onDeleteTag: (tagId: string) => void;
  className?: string;
};

export function TaskTagsPanel({
  tags,
  filterTagIds,
  isSaving,
  onToggleFilter,
  onClearFilters,
  onCreateTag,
  onDeleteTag,
  className,
}: TaskTagsPanelProps) {
  const [name, setName] = useState("");
  const [color, setColor] = useState<string>(TAG_COLOR_PRESETS[0]);

  const submit = (event: React.FormEvent) => {
    event.preventDefault();
    const trimmed = name.trim();
    if (!trimmed) {
      return;
    }
    onCreateTag({ name: trimmed, color });
    setName("");
  };

  return (
    <section className={cn(styles.root, className)} aria-label="Tags">
      <div className={styles.header}>
        <h2 className={styles.title}>Tags</h2>
        {filterTagIds.length > 0 ? (
          <button
            type="button"
            className={styles.clearBtn}
            onClick={onClearFilters}
          >
            Clear filter
          </button>
        ) : null}
      </div>

      {tags.length > 0 ? (
        <ul className={styles.tagList}>
          {tags.map((tag) => (
            <li key={tag.id} className={styles.tagRow}>
              <TaskTagPill
                tag={tag}
                selected={filterTagIds.includes(tag.id)}
                onClick={() => onToggleFilter(tag.id)}
              />
              <button
                type="button"
                className={styles.deleteBtn}
                aria-label={`Delete ${tag.name}`}
                disabled={isSaving}
                onClick={() => onDeleteTag(tag.id)}
              >
                <Icon name="trash" className="h-3 w-3" />
              </button>
            </li>
          ))}
        </ul>
      ) : (
        <p className={styles.hint}>No tags yet. Add one below.</p>
      )}

      <form className={styles.createForm} onSubmit={submit}>
        <input
          className={styles.nameInput}
          value={name}
          placeholder="New tag…"
          onChange={(event) => setName(event.target.value)}
        />
        <div className={styles.colors} role="listbox" aria-label="Tag color">
          {TAG_COLOR_PRESETS.map((preset) => (
            <button
              key={preset}
              type="button"
              role="option"
              aria-selected={color === preset}
              className={cn(
                styles.colorBtn,
                color === preset && styles.colorBtnOn,
              )}
              style={{ backgroundColor: preset }}
              aria-label={`Color ${preset}`}
              onClick={() => setColor(preset)}
            />
          ))}
        </div>
        <button
          type="submit"
          className={styles.addBtn}
          disabled={isSaving || !name.trim()}
        >
          Add tag
        </button>
      </form>
    </section>
  );
}
