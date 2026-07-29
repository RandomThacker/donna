"use client";

import { useEffect, useMemo, useState } from "react";

import { cn } from "@/lib/cn";

import { phoneFullscreenStyles as styles } from "./DashboardPhone.styles";
import type { DashboardPhoneFullscreenProps } from "./DashboardPhone.types";
import { IMessageChat } from "./IMessageChat";
import { IMessageList } from "./IMessageList";

export function DashboardPhoneFullscreen({
  phone,
  onClose,
  exiting = false,
  onCloseComplete,
}: DashboardPhoneFullscreenProps) {
  const [activeId, setActiveId] = useState<string | null>(null);

  const activeConversation = useMemo(
    () => phone.conversations.find((item) => item.id === activeId) ?? null,
    [activeId, phone.conversations],
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
            onBack={() => setActiveId(null)}
            onClose={onClose}
          />
        ) : (
          <IMessageList
            conversations={phone.conversations}
            onOpen={setActiveId}
            onClose={onClose}
          />
        )}
      </div>
    </div>
  );
}
