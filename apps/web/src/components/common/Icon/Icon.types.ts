export type IconName =
  | "sunrise"
  | "calendar"
  | "spark"
  | "check"
  | "arrow"
  | "google"
  | "microsoft"
  | "clock"
  | "moon"
  | "home"
  | "tasks"
  | "notes"
  | "memory"
  | "settings"
  | "link"
  | "user"
  | "mic"
  | "send"
  | "circle"
  | "chevronLeft"
  | "chevronRight"
  | "search"
  | "compose"
  | "plus"
  | "info"
  | "refresh"
  | "close"
  | "mapPin"
  | "pin"
  | "trash"
  | "repeat"
  | "sun";

export type IconProps = {
  name: IconName;
  className?: string;
};
