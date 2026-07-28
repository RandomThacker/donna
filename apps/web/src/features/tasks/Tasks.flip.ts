import { useLayoutEffect, useRef } from "react";

/** FLIP animation when list order changes (same keys, new positions). */
export function useFlipList(ids: string[], durationMs = 240) {
  const nodes = useRef(new Map<string, HTMLElement>());
  const previous = useRef(new Map<string, DOMRect>());

  useLayoutEffect(() => {
    const nextRects = new Map<string, DOMRect>();

    for (const id of ids) {
      const el = nodes.current.get(id);
      if (!el) {
        continue;
      }
      const next = el.getBoundingClientRect();
      nextRects.set(id, next);
      const first = previous.current.get(id);
      if (!first) {
        continue;
      }
      const dy = first.top - next.top;
      if (Math.abs(dy) < 1) {
        continue;
      }
      el.animate(
        [
          { transform: `translateY(${dy}px)` },
          { transform: "translateY(0)" },
        ],
        {
          duration: durationMs,
          easing: "cubic-bezier(0.22, 1, 0.36, 1)",
        },
      );
    }

    previous.current = nextRects;
  }, [ids, durationMs]);

  return (id: string) => (el: HTMLElement | null) => {
    if (el) {
      nodes.current.set(id, el);
    } else {
      nodes.current.delete(id);
    }
  };
}
