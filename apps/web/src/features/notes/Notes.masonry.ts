import { useEffect, useMemo, useState } from "react";

import type { Note } from "./Notes.types";

/** Rough height proxy so we can pack into the shortest column (Pinterest-style). */
export function estimateNoteHeight(note: Note): number {
  const titleLines = note.title ? 1 : 0;
  const contentLines = Math.max(1, Math.ceil(note.content.length / 42));
  const pin = note.pinned ? 18 : 0;
  return 56 + pin + titleLines * 22 + contentLines * 22;
}

export function masonryColumnCount(width: number): number {
  if (width < 520) {
    return 1;
  }
  if (width < 780) {
    return 2;
  }
  if (width < 1100) {
    return 3;
  }
  return 4;
}

/** Place each note into the currently shortest column. */
export function packNotesIntoColumns(notes: Note[], columnCount: number): Note[][] {
  const count = Math.max(1, columnCount);
  const columns: Note[][] = Array.from({ length: count }, () => []);
  const heights = Array.from({ length: count }, () => 0);

  for (const note of notes) {
    let shortest = 0;
    for (let i = 1; i < count; i += 1) {
      if ((heights[i] ?? 0) < (heights[shortest] ?? 0)) {
        shortest = i;
      }
    }
    columns[shortest]!.push(note);
    heights[shortest] = (heights[shortest] ?? 0) + estimateNoteHeight(note) + 16;
  }

  return columns;
}

export function useMasonryColumnCount(container: HTMLElement | null): number {
  const [count, setCount] = useState(1);

  useEffect(() => {
    if (!container) {
      return;
    }

    const update = () => {
      setCount(masonryColumnCount(container.clientWidth));
    };
    update();

    const observer = new ResizeObserver(update);
    observer.observe(container);
    return () => observer.disconnect();
  }, [container]);

  return count;
}

export function usePackedNoteColumns(
  notes: Note[],
  columnCount: number,
): Note[][] {
  return useMemo(
    () => packNotesIntoColumns(notes, columnCount),
    [notes, columnCount],
  );
}
