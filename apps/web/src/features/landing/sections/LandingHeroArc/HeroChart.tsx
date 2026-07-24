import type { HeroChartKind } from "../../Landing.types";

export function HeroChart({ kind }: { kind: HeroChartKind }) {
  if (kind === "line") {
    return (
      <svg viewBox="0 0 120 36" className="h-9 w-full" aria-hidden>
        <polyline
          points="0,30 20,24 40,27 60,14 80,18 100,7 120,10"
          fill="none"
          stroke="currentColor"
          strokeWidth={2}
          strokeLinecap="round"
          strokeLinejoin="round"
          className="text-donna-copper-bright"
        />
      </svg>
    );
  }

  const bars = [14, 22, 12, 28, 18, 32];

  return (
    <svg viewBox="0 0 120 36" className="h-9 w-full" aria-hidden>
      {bars.map((height, index) => (
        <rect
          key={index}
          x={index * 20 + 4}
          y={36 - height}
          width={10}
          height={height}
          rx={2}
          className="fill-donna-copper/70"
        />
      ))}
    </svg>
  );
}
