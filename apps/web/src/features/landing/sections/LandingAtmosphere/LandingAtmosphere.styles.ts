export const atmosphereStyles = {
  root: "pointer-events-none fixed inset-0 z-0 overflow-hidden",
  base: "absolute inset-0 bg-donna-bg",
  bloomTop: [
    "absolute left-1/2 top-[-30%] h-[70vmax] w-[70vmax] -translate-x-1/2 rounded-full",
    "bg-[radial-gradient(circle_at_center,var(--donna-accent-soft)_0%,transparent_62%)]",
    "blur-[50px]",
  ].join(" "),
  bloomLeft: [
    "absolute left-[-18%] top-[8%] h-[45vmax] w-[45vmax] rounded-full",
    "bg-[radial-gradient(circle_at_center,var(--donna-accent-soft)_0%,transparent_66%)]",
    "blur-[60px]",
  ].join(" "),
  bloomRight: "hidden",
  vignette: "absolute inset-0 bg-[radial-gradient(ellipse_at_center,transparent_40%,var(--donna-bg)_100%)] opacity-80",
  grain: [
    "absolute inset-0 opacity-[0.03] mix-blend-overlay",
    "bg-[url('data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 width=%27160%27 height=%27160%27%3E%3Cfilter id=%27n%27%3E%3CfeTurbulence type=%27fractalNoise%27 baseFrequency=%270.85%27 numOctaves=%274%27 stitchTiles=%27stitch%27/%3E%3C/filter%3E%3Crect width=%27160%27 height=%27160%27 filter=%27url(%23n)%27 opacity=%270.55%27/%3E%3C/svg%3E')]",
  ].join(" "),
} as const;
