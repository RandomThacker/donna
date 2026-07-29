"use client";

import { installPromptStyles as styles } from "./PwaInstallPrompt.styles";
import { usePwaInstall } from "./usePwaInstall";

export function PwaInstallPrompt() {
  const { visible, install, dismiss } = usePwaInstall();

  if (!visible) {
    return null;
  }

  return (
    <aside className={styles.banner} aria-live="polite">
      <div className={styles.row}>
        <span className={styles.mark} aria-hidden />
        <div className={styles.copy}>
          <p className={styles.title}>Install Donna</p>
          <p className={styles.body}>
            Pin Donna to your home screen for a calmer, app-like experience —
            same assistant, fewer browser tabs.
          </p>
          <div className={styles.actions}>
            <button type="button" className={styles.installBtn} onClick={() => void install()}>
              Install
            </button>
            <button type="button" className={styles.dismissBtn} onClick={dismiss}>
              Not now
            </button>
          </div>
        </div>
      </div>
    </aside>
  );
}
