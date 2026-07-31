"use client";

import { usePhoneThreadScreen } from "./usePhoneThreadScreen";
import { phoneStyles as styles } from "./DashboardPhone.styles";
import type { DashboardPhoneProps } from "./DashboardPhone.types";
import { IMessageChat } from "./IMessageChat";
import { IMessageList } from "./IMessageList";
import { StatusBarIcons } from "./StatusBarIcons";
import { StatusBarTime } from "./StatusBarTime";

export function DashboardPhone({ phone }: DashboardPhoneProps) {
  const screen = usePhoneThreadScreen(phone.conversations);

  return (
    <div id="phone" className={styles.wrap}>
      <div className={styles.frame}>
        <span className={styles.sideBtnLeft} aria-hidden />
        <span className={styles.sideBtnVol} aria-hidden />
        <span className={styles.sideBtnRight} aria-hidden />
        <div className={styles.screen}>
          <div className={styles.statusBar}>
            <StatusBarTime />
            <StatusBarIcons />
          </div>
          <div className={styles.island} aria-hidden />
          <div className={styles.content}>
            {screen.activeConversation ? (
              <IMessageChat
                conversation={screen.activeConversation}
                live={screen.activeConversation.id === "donna"}
                onBack={() => {
                  screen.setActiveId(null);
                  screen.donna.refresh();
                }}
              />
            ) : (
              <IMessageList
                conversations={screen.threads}
                onOpen={screen.setActiveId}
              />
            )}
          </div>
          <div className={styles.homeIndicator} aria-hidden />
        </div>
      </div>
    </div>
  );
}
