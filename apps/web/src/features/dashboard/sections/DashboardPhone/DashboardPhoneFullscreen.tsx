"use client";

import { useEffect } from "react";

import { cn } from "@/lib/cn";

import { usePhoneThreadScreen } from "./usePhoneThreadScreen";
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
  const screen = usePhoneThreadScreen(phone.conversations);

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
        {screen.activeConversation ? (
          <IMessageChat
            conversation={screen.activeConversation}
            live={screen.activeConversation.id === "donna"}
            onBack={() => {
              screen.setActiveId(null);
              screen.donna.refresh();
            }}
            onClose={onClose}
          />
        ) : (
          <IMessageList
            conversations={screen.threads}
            onOpen={screen.setActiveId}
            onClose={onClose}
          />
        )}
      </div>
    </div>
  );
}
