"use client";

import { useEffect, useRef } from "react";

import { Icon } from "@/components/common";
import { useChatSession } from "@/features/chat/Chat.logic";
import type { ChatMessage } from "@/features/chat/Chat.types";
import { cn } from "@/lib/cn";

import type { IMessageBubble } from "../../Dashboard.types";
import { BubbleTail } from "./BubbleTail";
import { iMessageStyles as styles } from "./DashboardPhone.styles";
import type { IMessageChatProps } from "./DashboardPhone.types";

function isLastInGroup(messages: IMessageBubble[], index: number) {
  const current = messages[index];
  const next = messages[index + 1];
  if (!current) return true;
  return !next || next.role !== current.role;
}

function isFirstInGroup(messages: IMessageBubble[], index: number) {
  const current = messages[index];
  const prev = messages[index - 1];
  if (!current) return true;
  return !prev || prev.role !== current.role;
}

function toBubbles(messages: ChatMessage[]): IMessageBubble[] {
  return messages.map((message) => ({
    id: message.id,
    role: message.role,
    text: message.text,
  }));
}

export function IMessageChat({
  conversation,
  onBack,
  onClose,
  live = false,
  initialDraft = "",
  showBack = true,
}: IMessageChatProps) {
  const session = useChatSession(live ? initialDraft : "", { enabled: live });
  const bottomRef = useRef<HTMLDivElement | null>(null);

  const bubbles: IMessageBubble[] = live
    ? toBubbles(session.messages)
    : conversation.messages;

  useEffect(() => {
    if (!live) return;
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [live, session.messages, session.sending]);

  const canSend =
    live &&
    session.draft.trim().length > 0 &&
    !session.sending &&
    !session.loadingHistory;

  return (
    <div className={styles.chatRoot}>
      <header className={styles.chatNav}>
        {showBack ? (
          <button type="button" className={styles.back} onClick={onBack}>
            <Icon name="chevronLeft" className="h-4 w-4" />
            <span>Messages</span>
          </button>
        ) : (
          <span className={styles.back} aria-hidden />
        )}
        <div className={styles.chatTitleWrap}>
          <span className={styles.chatAvatar}>{conversation.name.slice(0, 1)}</span>
          <span className={styles.chatName}>{conversation.name}</span>
        </div>
        {onClose ? (
          <button
            type="button"
            className={styles.chatInfo}
            aria-label="Close"
            onClick={onClose}
          >
            <Icon name="close" className="h-4 w-4" />
          </button>
        ) : (
          <button type="button" className={styles.chatInfo} aria-label="Info">
            <Icon name="info" className="h-4 w-4" />
          </button>
        )}
      </header>

      <div className={styles.chatBody}>
        <p className={styles.stamp}>
          {live
            ? "Donna"
            : `Today ${conversation.messages[0]?.time ?? ""}`}
        </p>
        {live && session.loadingHistory && bubbles.length === 0 ? (
          <p className={cn(styles.stamp, "mt-2")} aria-live="polite">
            Loading…
          </p>
        ) : null}
        {bubbles.map((message, index) => {
          const incoming = message.role === "donna";
          const last = isLastInGroup(bubbles, index);
          const first = isFirstInGroup(bubbles, index);

          return (
            <div
              key={message.id}
              className={cn(
                styles.bubbleWrap,
                incoming ? styles.bubbleWrapIn : styles.bubbleWrapOut,
                first ? styles.bubbleSpaced : styles.bubbleGrouped,
              )}
            >
              <div
                className={cn(
                  styles.bubble,
                  "whitespace-pre-wrap",
                  incoming ? styles.bubbleIn : styles.bubbleOut,
                  last
                    ? incoming
                      ? styles.bubbleInLast
                      : styles.bubbleOutLast
                    : incoming
                      ? styles.bubbleInMiddle
                      : styles.bubbleOutMiddle,
                )}
              >
                {message.text}
              </div>
              {last ? (
                <BubbleTail
                  side={incoming ? "in" : "out"}
                  className={incoming ? styles.tailIn : styles.tailOut}
                />
              ) : null}
            </div>
          );
        })}
        {live && session.sending ? (
          <p className={cn(styles.stamp, "mt-2")} aria-live="polite">
            Donna is typing…
          </p>
        ) : null}
        <div ref={bottomRef} />
      </div>

      <form
        className={styles.composer}
        onSubmit={(event) => {
          event.preventDefault();
          if (live) void session.send();
        }}
      >
        <button type="button" className={styles.plus} aria-label="Apps">
          <Icon name="plus" className="h-4 w-4" />
        </button>
        <div className={styles.inputShell}>
          <input
            type="text"
            className={styles.input}
            placeholder="iMessage"
            aria-label="iMessage"
            readOnly={!live}
            disabled={live && session.sending}
            value={live ? session.draft : ""}
            onChange={
              live
                ? (event) => session.setDraft(event.target.value)
                : undefined
            }
          />
          {canSend ? (
            <button
              type="submit"
              className={styles.send}
              aria-label="Send"
            >
              <Icon name="arrow" className="h-3.5 w-3.5 -rotate-90" />
            </button>
          ) : (
            <button type="button" className={styles.mic} aria-label="Dictation">
              <Icon name="mic" className="h-[18px] w-[18px]" />
            </button>
          )}
        </div>
      </form>
    </div>
  );
}
