"use client";

import { useCallback, useEffect, useRef, useState } from "react";

import { sendChatCommand } from "./Chat.api";
import type { ChatMessage } from "./Chat.types";

function newId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

const WELCOME: ChatMessage = {
  id: "welcome",
  role: "donna",
  text: "Hi — tell me what to do.\n\nTry: Add task Finish API\nOr: What do I have tomorrow?\nOr: What's due today?",
  createdAt: Date.now(),
};

export function useChatSession(initialDraft = "") {
  const [messages, setMessages] = useState<ChatMessage[]>([WELCOME]);
  const [draft, setDraft] = useState(initialDraft);
  const [sending, setSending] = useState(false);
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const appliedPrefill = useRef(false);

  useEffect(() => {
    if (appliedPrefill.current) return;
    if (initialDraft.trim()) {
      setDraft(initialDraft);
      appliedPrefill.current = true;
    }
  }, [initialDraft]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
  }, [messages, sending]);

  const send = useCallback(async () => {
    const text = draft.trim();
    if (!text || sending) {
      return;
    }
    const userMsg: ChatMessage = {
      id: newId(),
      role: "user",
      text,
      createdAt: Date.now(),
    };
    setMessages((prev) => [...prev, userMsg]);
    setDraft("");
    setSending(true);
    try {
      const result = await sendChatCommand(text);
      setMessages((prev) => [
        ...prev,
        {
          id: newId(),
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
  }, [draft, sending]);

  return {
    messages,
    draft,
    setDraft,
    sending,
    send,
    bottomRef,
  };
}
