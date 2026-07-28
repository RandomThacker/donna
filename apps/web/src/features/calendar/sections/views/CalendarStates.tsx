import { timelineStyles as styles } from "./view.styles";

export function CalendarSkeleton() {
  return (
    <div className={styles.skeleton} aria-hidden>
      <div className={`${styles.skeletonBlock} h-10 w-48`} />
      <div className={`${styles.skeletonBlock} h-24 w-full`} />
      <div className={`${styles.skeletonBlock} h-[28rem] w-full`} />
    </div>
  );
}

export function CalendarErrorState({
  message,
  onRetry,
}: {
  message: string;
  onRetry: () => void;
}) {
  return (
    <div className={styles.error} role="alert">
      <p className={styles.errorTitle}>Couldn’t load your calendar</p>
      <p className={styles.errorBody}>{message}</p>
      <button type="button" className={styles.retry} onClick={onRetry}>
        Try again
      </button>
    </div>
  );
}

export function CalendarEmptySources() {
  return (
    <div className={styles.error}>
      <p className={styles.errorTitle}>No calendars yet</p>
      <p className={styles.errorBody}>
        Connect a Google or Microsoft calendar from Integrations,
        then sync.
      </p>
      <a className={styles.retry} href="/dashboard/integrations">
        Open Integrations
      </a>
    </div>
  );
}
