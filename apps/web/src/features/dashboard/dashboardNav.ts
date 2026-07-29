import type { IconName } from "@/components/common";

export type DashboardNavItem = {
  id: string;
  label: string;
  icon: IconName;
  href: string;
};

export const dashboardNavItems: DashboardNavItem[] = [
  { id: "home", label: "Home", icon: "home", href: "/dashboard" },
  {
    id: "calendar",
    label: "Calendar",
    icon: "calendar",
    href: "/dashboard/calendar",
  },
  { id: "tasks", label: "Todo", icon: "tasks", href: "/dashboard/tasks" },
  { id: "notes", label: "Notes", icon: "notes", href: "/dashboard/notes" },
  {
    id: "memories",
    label: "Memories",
    icon: "memory",
    href: "/dashboard/memories",
  },
  {
    id: "integrations",
    label: "Integrations",
    icon: "link",
    href: "/dashboard/integrations",
  },
  {
    id: "settings",
    label: "Settings",
    icon: "settings",
    href: "/dashboard/settings",
  },
];

/** Mobile bottom bar — subset of primary nav. */
export const dashboardBottomNavIds = [
  "home",
  "calendar",
  "tasks",
  "notes",
  "settings",
] as const;

/** Active nav item for the current path — only one selection at a time. */
export function navItemsForPath(
  pathname: string,
): Array<DashboardNavItem & { active: boolean }> {
  const normalized =
    pathname.length > 1 && pathname.endsWith("/")
      ? pathname.slice(0, -1)
      : pathname;

  const activeId =
    normalized === "/dashboard"
      ? "home"
      : (dashboardNavItems.find(
          (item) =>
            item.id !== "home" &&
            (normalized === item.href ||
              normalized.startsWith(`${item.href}/`)),
        )?.id ?? null);

  return dashboardNavItems.map((item) => ({
    ...item,
    active: item.id === activeId,
  }));
}

export function bottomNavItemsForPath(
  pathname: string,
): Array<DashboardNavItem & { active: boolean }> {
  const ids: ReadonlySet<string> = new Set(dashboardBottomNavIds);
  return navItemsForPath(pathname).filter((item) => ids.has(item.id));
}
