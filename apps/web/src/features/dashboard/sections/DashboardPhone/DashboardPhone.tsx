"use client";

import { useMemo, useState } from "react";

import { useDonnaThreadSummary } from "@/features/chat/useDonnaThreadSummary";

import type { IMessageConversation } from "../../Dashboard.types";
import { phoneStyles as styles } from "./DashboardPhone.styles";
import type { DashboardPhoneProps } from "./DashboardPhone.types";
import { IMessageChat } from "./IMessageChat";
import { IMessageList } from "./IMessageList";
import { StatusBarIcons } from "./StatusBarIcons";
import { StatusBarTime } from "./StatusBarTime";

function withDonnaSummary(
  conversations: IMessageConversation[],
  preview: string,
  time: string,
  unread: number,
): IMessageConversation[] {
  return conversations.map((conversation) =>
    conversation.id === "donna"
      ? {
          ...conversation,
          preview,
          time: time || conversation.time,
          unread,
        }
      : conversation,
  );
}

export function DashboardPhone({ phone }: DashboardPhoneProps) {
  const [activeId, setActiveId] = useState<string | null>("donna");
  const donna = useDonnaThreadSummary();

  const conversations = useMemo(
    () => withDonnaSummary(phone.conversations, donna.preview, donna.time, donna.unread),
    [phone.conversations, donna.preview, donna.time, donna.unread],
  );

  const activeConversation = useMemo(
    () => conversations.find((item) => item.id === activeId) ?? null,
    [activeId, conversations],
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
                onBack={() => {
                  setActiveId(null);
                  donna.refresh();
                }}
              />
            ) : (
              <IMessageList
                conversations={conversations}
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
