import type { DashboardNavItem } from "../../Dashboard.types";

export type DashboardSidebarProps = {
  items: DashboardNavItem[];
  profileName: string;
  profileInitials: string;
  profileAvatarUrl?: string | null;
  onSignOut?: () => void;
};
