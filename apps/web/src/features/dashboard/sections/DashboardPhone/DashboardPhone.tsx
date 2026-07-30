"use client";

import { useMemo, useState } from "react";

import { phoneStyles as styles } from "./DashboardPhone.styles";
import type { DashboardPhoneProps } from "./DashboardPhone.types";
import { IMessageChat } from "./IMessageChat";
import { IMessageList } from "./IMessageList";
import { StatusBarIcons } from "./StatusBarIcons";
import { StatusBarTime } from "./StatusBarTime";

export function DashboardPhone({ phone }: DashboardPhoneProps) {
  const [activeId, setActiveId] = useState<string | null>(null);

  const activeConversation = useMemo(
    () => phone.conversations.find((item) => item.id === activeId) ?? null,
    [activeId, phone.conversations],
  );

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
            {activeConversation ? (
              <IMessageChat
                conversation={activeConversation}
                live={activeConversation.id === "donna"}
                onBack={() => setActiveId(null)}
              />
            ) : (
              <IMessageList
                conversations={phone.conversations}
                onOpen={setActiveId}
              />
            )}
          </div>
          <div className={styles.homeIndicator} aria-hidden />
        </div>
      </div>
    </div>
  );
}
