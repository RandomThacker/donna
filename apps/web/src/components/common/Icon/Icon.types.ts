export type IconName =
  | "sunrise"
  | "calendar"
  | "spark"
  | "check"
  | "arrow"
  | "google"
  | "clock"
  | "moon"
  | "home"
  | "tasks"
  | "notes"
  | "memory"
  | "settings"
  | "user"
  | "mic"
  | "send"
  | "circle"
  | "chevronLeft"
  | "chevronRight"
  | "search"
  | "compose"
  | "plus"
  | "info";

export type IconProps = {
  name: IconName;
  className?: string;
};
