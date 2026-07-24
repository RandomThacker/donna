import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { BentoBox } from "../BentoBox";
import { quickTasksStyles as styles } from "./DashboardQuickTasks.styles";
import type { DashboardQuickTasksProps } from "./DashboardQuickTasks.types";

export function DashboardQuickTasks({ tasks }: DashboardQuickTasksProps) {
  return (
    <BentoBox className={styles.box} title="Quick tasks">
      <ul className={styles.list}>
        {tasks.map((task) => (
          <li key={task.id}>
            <button type="button" className={styles.item} aria-pressed={task.done}>
              <span className={cn(styles.check, task.done && styles.checkDone)}>
                <Icon name="check" className="h-2.5 w-2.5" />
              </span>
              <span className={cn(styles.labelText, task.done && styles.labelDone)}>
                {task.label}
              </span>
            </button>
          </li>
        ))}
      </ul>
    </BentoBox>
  );
}
