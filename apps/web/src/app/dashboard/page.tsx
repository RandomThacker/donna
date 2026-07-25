import { RequireAuth } from "@/features/auth";
import { Dashboard } from "@/features/dashboard";

export default function DashboardPage() {
  return (
    <RequireAuth>
      <Dashboard />
    </RequireAuth>
  );
}
