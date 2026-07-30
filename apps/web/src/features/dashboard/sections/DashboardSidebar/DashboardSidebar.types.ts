import type { DashboardNavItem } from "../../Dashboard.types";

export type DashboardSidebarProps = {
  items: DashboardNavItem[];
  profileName: string;
  profileInitials: string;
  profileEmail?: string | null;
  profileAvatarUrl?: string | null;
  /** Optional unread/count badges keyed by nav item id. */
  badgeByNavId?: Partial<Record<string, number>>;
};
