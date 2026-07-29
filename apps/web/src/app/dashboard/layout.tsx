import type { ReactNode } from "react";

import { DashboardBottomBar } from "@/features/dashboard/sections/DashboardBottomBar";
import { DonnaPhoneFab } from "@/features/dashboard/sections/DashboardPhone";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <>
      {children}
      <DonnaPhoneFab />
      <DashboardBottomBar />
    </>
  );
}
