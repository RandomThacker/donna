"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { fetchChatMessages, sendChatCommand } from "./Chat.api";
import type { ChatHistoryMessage, ChatMessage } from "./Chat.types";

function newId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

const WELCOME: ChatMessage = {
  id: "welcome",
  role: "donna",
  text: "Hi — tell me what to do.\n\nTry: Add task Finish API\nOr: What do I have tomorrow?\nOr: What's due today?",
  createdAt: Date.now(),
};

function mapHistoryMessage(m: ChatHistoryMessage): ChatMessage {
  return {
    id: m.public_id || m.id,
    role: m.role === "user" ? "user" : "donna",
    text: m.content,
    createdAt: Date.parse(m.created_at) || Date.now(),
  };
}

export function useChatSession(
  initialDraft = "",
  options: { enabled?: boolean } = {},
) {
  const enabled = options.enabled !== false;
  const [messages, setMessages] = useState<ChatMessage[]>(
    enabled ? [] : [WELCOME],
  );
  const [draft, setDraft] = useState(initialDraft);
  const [sending, setSending] = useState(false);
  const [loadingHistory, setLoadingHistory] = useState(enabled);
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const appliedPrefill = useRef(false);

  useEffect(() => {
    if (!enabled) {
      setLoadingHistory(false);
      return;
    }
    let cancelled = false;
    (async () => {
      try {
        const history = await fetchChatMessages();
        if (cancelled) return;
        if (history.messages.length === 0) {
          setMessages([WELCOME]);
        } else {
          setMessages(history.messages.map(mapHistoryMessage));
        }
      } catch {
        if (!cancelled) {
          setMessages([WELCOME]);
        }
      } finally {
        if (!cancelled) {
          setLoadingHistory(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  useEffect(() => {
    if (appliedPrefill.current) return;
    if (initialDraft.trim()) {
      setDraft(initialDraft);
      appliedPrefill.current = true;
    }
  }, [initialDraft]);

  useEffect(() => {
    if (loadingHistory) return;
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, sending, loadingHistory]);

  const send = useCallback(async () => {
    const text = draft.trim();
    if (!text || sending || loadingHistory) {
      return;
    }
    const clientId = newId();
    const userMsg: ChatMessage = {
      id: clientId,
      role: "user",
      text,
      createdAt: Date.now(),
    };
    setMessages((prev) => {
      const withoutWelcome = prev.filter((m) => m.id !== "welcome");
      return [...withoutWelcome, userMsg];
    });
    setDraft("");
    setSending(true);
    try {
      const result = await sendChatCommand(text, clientId);
      setMessages((prev) => [
        ...prev.map((m) =>
          m.id === clientId && result.user_message_public_id
            ? { ...m, id: result.user_message_public_id }
            : m,
        ),
        {
          id: result.reply_message_public_id || newId(),
          role: "donna",
          text: result.reply,
          createdAt: Date.now(),
        },
      ]);
    } catch {
      setMessages((prev) => [
        ...prev,
        {
          id: newId(),
          role: "donna",
          text: "Something went wrong on my end. Try again in a moment.",
          createdAt: Date.now(),
        },
      ]);
    } finally {
      setSending(false);
    }
  }, [draft, sending, loadingHistory]);

  return {
    messages,
    draft,
    setDraft,
    sending,
    loadingHistory,
    send,
    bottomRef,
  };
}
