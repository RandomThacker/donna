import { getDashboardContent } from "./Dashboard.logic";
import { dashboardStyles as styles } from "./Dashboard.styles";
import { DashboardFocus } from "./sections/DashboardFocus";
import { DashboardGreeting } from "./sections/DashboardGreeting";
import { DashboardInsights } from "./sections/DashboardInsights";
import { DashboardPhone } from "./sections/DashboardPhone";
import { DashboardQuickTasks } from "./sections/DashboardQuickTasks";
import { DashboardSidebar } from "./sections/DashboardSidebar";
import { DashboardTimeline } from "./sections/DashboardTimeline";

export function Dashboard() {
  const { data } = getDashboardContent();

  return (
    <div className={styles.page}>
      <div className={styles.shell}>
        <DashboardSidebar
          items={data.nav}
          profileName={data.profileName}
          profileInitials={data.profileInitials}
        />
        <main className={styles.workspace}>
          <div className={styles.workspaceInner}>
            <div className={styles.bento}>
              <DashboardGreeting greeting={data.greeting} />
              <DashboardFocus focus={data.focus} />
              <DashboardInsights insights={data.insights} />
              <DashboardTimeline items={data.timeline} />
              <DashboardQuickTasks tasks={data.tasks} />
            </div>
            <div className={styles.phoneMobile}>
              <DashboardPhone phone={data.phone} />
            </div>
          </div>
        </main>
        <aside className={styles.phoneColumn} aria-label="Donna phone">
          <DashboardPhone phone={data.phone} />
        </aside>
      </div>
    </div>
  );
}
