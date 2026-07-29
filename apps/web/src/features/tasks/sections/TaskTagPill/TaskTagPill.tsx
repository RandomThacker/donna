import { cn } from "@/lib/cn";

import type { TaskTag } from "../../Tasks.types";
import { tagPillStyles as styles } from "./TaskTagPill.styles";

type TaskTagPillProps = {
  tag: TaskTag;
  className?: string;
  onClick?: () => void;
  selected?: boolean;
};

export function TaskTagPill({
  tag,
  className,
  onClick,
  selected,
}: TaskTagPillProps) {
  const style = {
    backgroundColor: `${tag.color}22`,
    borderColor: `${tag.color}66`,
    color: tag.color,
  };

  if (onClick) {
    return (
      <button
        type="button"
        className={cn(styles.pill, selected && styles.pillSelected, className)}
        style={style}
        onClick={onClick}
        aria-pressed={selected}
      >
        {tag.name}
      </button>
    );
  }

  return (
    <span className={cn(styles.pill, className)} style={style}>
      {tag.name}
    </span>
  );
}
