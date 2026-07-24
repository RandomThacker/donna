export const atmosphereStyles = {
  root: "pointer-events-none fixed inset-0 z-0 overflow-hidden",
  base: "absolute inset-0 bg-donna-void",
  bloomTop: [
    "absolute left-1/2 top-[-32%] h-[80vmax] w-[80vmax] -translate-x-1/2 rounded-full",
    "bg-[radial-gradient(circle_at_center,rgb(236_211_168_/_0.16)_0%,rgb(169_127_79_/_0.12)_28%,transparent_62%)]",
    "blur-[60px] animate-donna-drift",
  ].join(" "),
  bloomLeft: [
    "absolute left-[-18%] top-[6%] h-[52vmax] w-[52vmax] rounded-full",
    "bg-[radial-gradient(circle_at_center,rgb(203_169_125_/_0.14)_0%,transparent_66%)]",
    "blur-[70px] animate-donna-drift [animation-delay:-6s]",
  ].join(" "),
  bloomRight: [
    "absolute bottom-[-24%] right-[-14%] h-[58vmax] w-[58vmax] rounded-full",
    "bg-[radial-gradient(circle_at_center,rgb(61_40_28_/_0.6)_0%,transparent_68%)]",
    "blur-[90px]",
  ].join(" "),
  vignette: [
    "absolute inset-0",
    "bg-[radial-gradient(ellipse_at_center,transparent_0%,rgb(8_6_5_/_0.4)_52%,rgb(8_6_5_/_0.92)_100%)]",
  ].join(" "),
  grain: [
    "absolute inset-0 opacity-[0.04] mix-blend-overlay",
    "bg-[url('data:image/svg+xml,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 width=%27160%27 height=%27160%27%3E%3Cfilter id=%27n%27%3E%3CfeTurbulence type=%27fractalNoise%27 baseFrequency=%270.85%27 numOctaves=%274%27 stitchTiles=%27stitch%27/%3E%3C/filter%3E%3Crect width=%27160%27 height=%27160%27 filter=%27url(%23n)%27 opacity=%270.55%27/%3E%3C/svg%3E')]",
  ].join(" "),
} as const;
