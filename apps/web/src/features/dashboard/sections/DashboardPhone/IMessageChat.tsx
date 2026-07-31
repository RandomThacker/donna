"use client";

import { useEffect, useLayoutEffect, useRef, useState } from "react";

import { Icon } from "@/components/common";
import { useChatSession } from "@/features/chat/Chat.logic";
import type { ChatMessage } from "@/features/chat/Chat.types";
import {
  chatIntentSuggestions,
  pickSuggestionPhrase,
} from "@/features/chat/chatIntentSuggestions";
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

function formatMessageTime(createdAt?: number, fallback?: string): string {
  if (fallback?.trim()) {
    return fallback.trim();
  }
  if (!createdAt) {
    return "";
  }
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
  }).format(new Date(createdAt));
}

function toBubbles(messages: ChatMessage[]): IMessageBubble[] {
  return messages.map((message) => ({
    id: message.id,
    role: message.role,
    text: message.text,
    time: formatMessageTime(message.createdAt),
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
  const session = useChatSession(live ? initialDraft : "", {
    enabled: live,
    unreadOnOpen: live ? conversation.unread : 0,
  });
  const bodyRef = useRef<HTMLDivElement | null>(null);
  const inputRef = useRef<HTMLInputElement | null>(null);
  const historySeededRef = useRef(false);
  const seenIdsRef = useRef<Set<string>>(new Set());
  const freshThisRenderRef = useRef<string[]>([]);
  const [enterIds, setEnterIds] = useState<Set<string>>(() => new Set());
  const [typingPhase, setTypingPhase] = useState<"hidden" | "in" | "out">(
    "hidden",
  );

  const bubbles: IMessageBubble[] = live
    ? toBubbles(session.messages)
    : conversation.messages;
  const bubbleSignature = bubbles.map((message) => message.id).join("\0");
  const scrollSignature = `${bubbleSignature}|${typingPhase}|${session.newMessageBeforeId ?? ""}`;

  useEffect(() => {
    if (!live) {
      setTypingPhase("hidden");
      return;
    }
    if (session.sending) {
      setTypingPhase("in");
      return;
    }
    setTypingPhase((phase) => {
      if (phase !== "in") {
        return phase;
      }
      return "out";
    });
  }, [live, session.sending]);

  useEffect(() => {
    if (typingPhase !== "out") {
      return;
    }
    const timer = window.setTimeout(() => setTypingPhase("hidden"), 280);
    return () => window.clearTimeout(timer);
  }, [typingPhase]);

  freshThisRenderRef.current = [];
  if (!(live && session.loadingHistory)) {
    if (!historySeededRef.current) {
      for (const message of bubbles) {
        seenIdsRef.current.add(message.id);
      }
      historySeededRef.current = true;
    } else {
      for (const message of bubbles) {
        if (!seenIdsRef.current.has(message.id)) {
          freshThisRenderRef.current.push(message.id);
          seenIdsRef.current.add(message.id);
        }
      }
    }
  }
  const freshThisRender = freshThisRenderRef.current;
  const freshKey = freshThisRender.join("\0");

  useLayoutEffect(() => {
    if (!freshKey) {
      return;
    }
    const fresh = freshKey.split("\0");
    setEnterIds((prev) => {
      const next = new Set(prev);
      for (const id of fresh) next.add(id);
      return next;
    });
    window.setTimeout(() => {
      setEnterIds((prev) => {
        const next = new Set(prev);
        for (const id of fresh) next.delete(id);
        return next;
      });
    }, 700);
  }, [freshKey]);

  useLayoutEffect(() => {
    if (!live) return;
    const el = bodyRef.current;
    if (!el) return;
    const pin = () => {
      el.scrollTop = el.scrollHeight;
    };
    pin();
    const raf = window.requestAnimationFrame(pin);
    return () => window.cancelAnimationFrame(raf);
  }, [live, scrollSignature, session.loadingHistory]);

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

      <div ref={bodyRef} className={styles.chatBody}>
        <div className={styles.chatBodyInner}>
          <p className={styles.stamp}>{live ? "iMessage" : "Today"}</p>
          {live && session.loadingHistory && bubbles.length === 0 ? (
            <p className={cn(styles.stamp, "mt-2")} aria-live="polite">
              Loading…
            </p>
          ) : null}
          {bubbles.map((message, index) => {
            const incoming = message.role === "donna";
            const last = isLastInGroup(bubbles, index);
            const first = isFirstInGroup(bubbles, index);
            const timeLabel = formatMessageTime(undefined, message.time);
            const showNewDivider =
              live &&
              Boolean(session.newMessageBeforeId) &&
              message.id === session.newMessageBeforeId;
            const entering =
              enterIds.has(message.id) ||
              freshThisRender.includes(message.id);

            return (
              <div key={message.id} className="flex w-full flex-col">
                {showNewDivider ? (
                  <div className={styles.newMessageRule} role="status">
                    <span className={styles.newMessageLine} aria-hidden />
                    <span className={styles.newMessageLabel}>New Message</span>
                    <span className={styles.newMessageLine} aria-hidden />
                  </div>
                ) : null}
                <div
                  className={cn(
                    styles.bubbleWrap,
                    incoming ? styles.bubbleWrapIn : styles.bubbleWrapOut,
                    first && !showNewDivider
                      ? styles.bubbleSpaced
                      : styles.bubbleGrouped,
                    showNewDivider && styles.bubbleSpaced,
                    entering &&
                      (incoming ? styles.bubbleEnterIn : styles.bubbleEnterOut),
                  )}
                >
                  <div className="relative">
                    <div
                      className={cn(
                        styles.bubble,
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
                      <span className={styles.bubbleText}>{message.text}</span>
                      {timeLabel ? (
                        <span
                          className={cn(
                            styles.bubbleTime,
                            incoming ? styles.bubbleTimeIn : styles.bubbleTimeOut,
                          )}
                        >
                          {timeLabel}
                        </span>
                      ) : null}
                    </div>
                    {last ? (
                      <BubbleTail
                        side={incoming ? "in" : "out"}
                        className={incoming ? styles.tailIn : styles.tailOut}
                      />
                    ) : null}
                  </div>
                </div>
              </div>
            );
          })}
          {live && typingPhase !== "hidden" ? (
            <div
              className={cn(
                styles.typingWrap,
                typingPhase === "out" ? styles.typingExit : styles.typingEnter,
              )}
              aria-live="polite"
              aria-label="Donna is typing"
            >
              <div className="relative">
                <div className={styles.typingBubble}>
                  <span
                    className={styles.typingDot}
                    style={{ animationDelay: "0ms" }}
                  />
                  <span
                    className={styles.typingDot}
                    style={{ animationDelay: "140ms" }}
                  />
                  <span
                    className={styles.typingDot}
                    style={{ animationDelay: "280ms" }}
                  />
                </div>
                <BubbleTail side="in" className={styles.tailIn} />
              </div>
            </div>
          ) : null}
        </div>
      </div>

      <div className={styles.composerDock}>
        {live ? (
          <div
            className={styles.suggestionRow}
            role="list"
            aria-label="Suggested commands"
          >
            {chatIntentSuggestions.map((suggestion) => (
              <button
                key={suggestion.id}
                type="button"
                role="listitem"
                className={styles.suggestionPill}
                disabled={session.sending}
                onClick={() => {
                  session.setDraft(pickSuggestionPhrase(suggestion));
                  inputRef.current?.focus();
                }}
              >
                {suggestion.label}
              </button>
            ))}
          </div>
        ) : null}
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
              ref={inputRef}
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
              <button type="submit" className={styles.send} aria-label="Send">
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
    </div>
  );
}
