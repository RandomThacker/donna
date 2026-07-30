"use client";

import { useEffect, useMemo, useState } from "react";

import { useDonnaThreadSummary } from "@/features/chat/useDonnaThreadSummary";
import { cn } from "@/lib/cn";

import type { IMessageConversation } from "../../Dashboard.types";
import { phoneFullscreenStyles as styles } from "./DashboardPhone.styles";
import type { DashboardPhoneFullscreenProps } from "./DashboardPhone.types";
import { IMessageChat } from "./IMessageChat";
import { IMessageList } from "./IMessageList";

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

export function DashboardPhoneFullscreen({
  phone,
  onClose,
  exiting = false,
  onCloseComplete,
}: DashboardPhoneFullscreenProps) {
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

  useEffect(() => {
    const previous = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.body.style.overflow = previous;
    };
  }, []);

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onClose();
      }
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onClose]);

  return (
    <div
      className={cn(
        styles.root,
        exiting ? styles.rootExit : styles.rootEnter,
      )}
      role="dialog"
      aria-modal="true"
      aria-label="Donna messages"
      onAnimationEnd={() => {
        if (exiting) {
          onCloseComplete?.();
        }
      }}
    >
      <div className={styles.body}>
        {activeConversation ? (
          <IMessageChat
            conversation={activeConversation}
            live={activeConversation.id === "donna"}
            onBack={() => {
              setActiveId(null);
              donna.refresh();
            }}
            onClose={onClose}
          />
        ) : (
          <IMessageList
            conversations={conversations}
            onOpen={setActiveId}
            onClose={onClose}
          />
        )}
      </div>
    </div>
  );
}
