import { cn } from "@/lib/cn";

type BubbleTailProps = {
  side: "in" | "out";
  className?: string;
};

/** Classic iMessage bubble tail. */
export function BubbleTail({ side, className }: BubbleTailProps) {
  if (side === "in") {
    return (
      <svg
        viewBox="0 0 12 16"
        className={className}
        aria-hidden
        fill="currentColor"
      >
        <path d="M12 16C8 16 3.5 13.5 0 8.5 3.2 10.2 6.5 11 12 11V16Z" />
      </svg>
    );
  }

  return (
    <svg
      viewBox="0 0 12 16"
      className={cn(className, "scale-x-[-1]")}
      aria-hidden
      fill="currentColor"
    >
      <path d="M12 16C8 16 3.5 13.5 0 8.5 3.2 10.2 6.5 11 12 11V16Z" />
    </svg>
  );
}
