import type { ReactNode } from "react";

import { DashboardBottomBar } from "@/features/dashboard/sections/DashboardBottomBar";
import { DonnaPhoneFab } from "@/features/dashboard/sections/DashboardPhone";
import { WebPushRegister } from "@/features/notifications/WebPushRegister";

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <WebPushRegister />
      {children}
      <DonnaPhoneFab />
      <DashboardBottomBar />
    </>
  );
}
