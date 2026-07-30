"use client";

import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import { useChatSession } from "./Chat.logic";
import { chatStyles as styles } from "./Chat.styles";

type ChatThreadProps = {
  className?: string;
  compact?: boolean;
  initialDraft?: string;
};

/** Session-only command chat thread (no persistence). */
export function ChatThread({
  className,
  compact = false,
  initialDraft = "",
}: ChatThreadProps) {
  const { messages, draft, setDraft, sending, send, bottomRef } =
    useChatSession(initialDraft);

  return (
    <div className={cn("flex min-h-0 flex-1 flex-col", className)}>
      <div className={cn(styles.thread, compact && "px-3 py-3")}>
        {messages.map((message) => (
          <div
            key={message.id}
            className={message.role === "user" ? styles.rowUser : styles.rowDonna}
          >
            <div
              className={
                message.role === "user" ? styles.bubbleUser : styles.bubbleDonna
              }
            >
              {message.text}
            </div>
          </div>
        ))}
        {sending ? (
          <div className={styles.rowDonna}>
            <p className={styles.typing} aria-live="polite">
              Donna is typing…
            </p>
          </div>
        ) : null}
        <div ref={bottomRef} />
      </div>

      <div className={cn(styles.composer, compact && "px-3 py-2")}>
        <form
          className={styles.composerRow}
          onSubmit={(event) => {
            event.preventDefault();
            void send();
          }}
        >
          <textarea
            className={styles.input}
            rows={compact ? 1 : 2}
            value={draft}
            placeholder="Tell Donna what to do…"
            aria-label="Message Donna"
            disabled={sending}
            onChange={(event) => setDraft(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === "Enter" && !event.shiftKey) {
                event.preventDefault();
                void send();
              }
            }}
          />
          <button
            type="submit"
            className={styles.send}
            disabled={sending || !draft.trim()}
            aria-label="Send"
          >
            <Icon name="send" className="h-4 w-4" />
          </button>
        </form>
      </div>
    </div>
  );
}
