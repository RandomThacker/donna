import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type CSSProperties,
  type PointerEvent as ReactPointerEvent,
} from "react";

const FAB_SIZE = 48;
const EDGE_GAP = 16;
const DRAG_THRESHOLD = 8;
const BOTTOM_NAV_OFFSET = 82;

type FabPoint = { x: number; y: number };

function clampFabPoint(point: FabPoint): FabPoint {
  const maxX = Math.max(EDGE_GAP, window.innerWidth - FAB_SIZE - EDGE_GAP);
  // Keep the FAB above the bottom nav so it never steals tab taps.
  const maxY = Math.max(
    EDGE_GAP,
    window.innerHeight - FAB_SIZE - EDGE_GAP - BOTTOM_NAV_OFFSET,
  );
  return {
    x: Math.min(maxX, Math.max(EDGE_GAP, point.x)),
    y: Math.min(maxY, Math.max(EDGE_GAP, point.y)),
  };
}

export function useDonnaPhoneFab() {
  const [open, setOpen] = useState(false);
  const [exiting, setExiting] = useState(false);
  const [position, setPosition] = useState<FabPoint | null>(null);
  const [dragging, setDragging] = useState(false);

  const dragRef = useRef<{
    pointerId: number;
    startX: number;
    startY: number;
    originX: number;
    originY: number;
    moved: boolean;
  } | null>(null);

  useEffect(() => {
    function onResize() {
      setPosition((current) =>
        current ? clampFabPoint(current) : current,
      );
    }
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  const openPhone = useCallback(() => {
    setExiting(false);
    setOpen(true);
  }, []);

  const requestClose = useCallback(() => {
    setExiting(true);
  }, []);

  const finishClose = useCallback(() => {
    setOpen(false);
    setExiting(false);
  }, []);

  const onPointerDown = useCallback(
    (event: ReactPointerEvent<HTMLButtonElement>) => {
      if (event.button !== 0) {
        return;
      }
      const rect = event.currentTarget.getBoundingClientRect();
      const origin = position ?? { x: rect.left, y: rect.top };
      if (!position) {
        setPosition(origin);
      }
      dragRef.current = {
        pointerId: event.pointerId,
        startX: event.clientX,
        startY: event.clientY,
        originX: origin.x,
        originY: origin.y,
        moved: false,
      };
      event.currentTarget.setPointerCapture(event.pointerId);
    },
    [position],
  );

  const onPointerMove = useCallback(
    (event: ReactPointerEvent<HTMLButtonElement>) => {
      const drag = dragRef.current;
      if (!drag || drag.pointerId !== event.pointerId) {
        return;
      }
      const dx = event.clientX - drag.startX;
      const dy = event.clientY - drag.startY;
      if (!drag.moved && Math.hypot(dx, dy) < DRAG_THRESHOLD) {
        return;
      }
      drag.moved = true;
      setDragging(true);
      setPosition(
        clampFabPoint({
          x: drag.originX + dx,
          y: drag.originY + dy,
        }),
      );
    },
    [],
  );

  const onPointerUp = useCallback(
    (event: ReactPointerEvent<HTMLButtonElement>) => {
      const drag = dragRef.current;
      if (!drag || drag.pointerId !== event.pointerId) {
        return;
      }
      const wasDrag = drag.moved;
      dragRef.current = null;
      setDragging(false);
      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
      if (!wasDrag) {
        openPhone();
      }
    },
    [openPhone],
  );

  const fabStyle: CSSProperties =
    position != null
      ? {
          left: position.x,
          top: position.y,
          right: "auto",
          bottom: "auto",
        }
      : {
          right: EDGE_GAP,
          bottom: `calc(${BOTTOM_NAV_OFFSET}px + env(safe-area-inset-bottom))`,
        };

  return {
    open,
    exiting,
    dragging,
    fabStyle,
    onPointerDown,
    onPointerMove,
    onPointerUp,
    requestClose,
    finishClose,
  };
}
