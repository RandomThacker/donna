"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { fetchChatMessages, sendChatCommand } from "./Chat.api";
import type { ChatHistoryMessage, ChatMessage } from "./Chat.types";
import {
  playChatReceiveSound,
  playChatSendSound,
  setLiveChatOpen,
} from "./chatSounds";
import { chatSummaryQueryKey } from "./useDonnaThreadSummary";

function newId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

const WELCOME: ChatMessage = {
  id: "welcome",
  role: "donna",
  text: "Hi — tell me what to do.\n\nTry: Add task Finish API\nOr: What do I have tomorrow?\nOr: What's due today?",
  createdAt: Date.now(),
};

const POLL_MS = 12_000;
const TYPING_MIN_MS = 1_000;
const TYPING_EXIT_MS = 280;

function mapHistoryMessage(m: ChatHistoryMessage): ChatMessage {
  return {
    id: m.public_id || m.id,
    role: m.role === "user" ? "user" : "donna",
    text: m.content,
    createdAt: Date.parse(m.created_at) || Date.now(),
  };
}

function wait(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

/** Keep first occurrence of each id — React keys must stay unique. */
function dedupeById(messages: ChatMessage[]): ChatMessage[] {
  const seen = new Set<string>();
  const out: ChatMessage[] = [];
  for (const message of messages) {
    if (seen.has(message.id)) {
      continue;
    }
    seen.add(message.id);
    out.push(message);
  }
  return out;
}

/** Oldest Donna message among the trailing `unread` assistant messages. */
function firstUnreadDonnaId(
  messages: ChatMessage[],
  unread: number,
): string | null {
  if (unread <= 0) {
    return null;
  }
  let remaining = unread;
  let firstId: string | null = null;
  for (let i = messages.length - 1; i >= 0; i -= 1) {
    const message = messages[i];
    if (!message || message.role !== "donna") {
      continue;
    }
    firstId = message.id;
    remaining -= 1;
    if (remaining <= 0) {
      break;
    }
  }
  return firstId;
}

export function useChatSession(
  initialDraft = "",
  options: { enabled?: boolean; unreadOnOpen?: number } = {},
) {
  const enabled = options.enabled !== false;
  const unreadOnOpenRef = useRef(Math.max(0, options.unreadOnOpen ?? 0));
  const queryClient = useQueryClient();
  const [messages, setMessages] = useState<ChatMessage[]>(
    enabled ? [] : [WELCOME],
  );
  const [draft, setDraft] = useState(initialDraft);
  const [sending, setSending] = useState(false);
  const [loadingHistory, setLoadingHistory] = useState(enabled);
  /** First Donna message id in an unread / newly arrived batch. */
  const [newMessageBeforeId, setNewMessageBeforeId] = useState<string | null>(
    null,
  );
  const bottomRef = useRef<HTMLDivElement | null>(null);
  const appliedPrefill = useRef(false);
  const knownIdsRef = useRef<Set<string>>(new Set());
  const historyReadyRef = useRef(false);
  const dividerSetRef = useRef(false);

  const refreshSummary = useCallback(() => {
    void queryClient.invalidateQueries({ queryKey: chatSummaryQueryKey });
  }, [queryClient]);

  const rememberIds = useCallback((items: ChatMessage[]) => {
    for (const item of items) {
      knownIdsRef.current.add(item.id);
    }
  }, []);

  const mergeRemoteMessages = useCallback(
    (remote: ChatMessage[]) => {
      const arrivals = remote.filter((message) => !knownIdsRef.current.has(message.id));
      for (const message of remote) {
        knownIdsRef.current.add(message.id);
      }

      if (
        historyReadyRef.current &&
        !dividerSetRef.current
      ) {
        const firstDonna = arrivals.find((m) => m.role === "donna");
        if (firstDonna) {
          dividerSetRef.current = true;
          setNewMessageBeforeId(firstDonna.id);
        }
      }

      if (
        historyReadyRef.current &&
        arrivals.some((message) => message.role === "donna")
      ) {
        playChatReceiveSound();
      }

      if (arrivals.length === 0 && remote.length === 0) {
        return;
      }

      setMessages((prev) => {
        const byId = new Map(prev.map((m) => [m.id, m]));
        for (const message of remote) {
          byId.set(message.id, message);
        }

        const welcomeOnly =
          prev.length === 1 && prev[0]?.id === "welcome" && remote.length > 0;
        if (welcomeOnly) {
          return dedupeById(remote);
        }

        // Remote ids win order. Drop local-only rows that the server already
        // echoed under a different id (optimistic clientId still in prev while
        // public_id is in remote) by preferring unique ids only.
        const remoteIds = new Set(remote.map((m) => m.id));
        const order = [...remoteIds];
        const extras = prev.filter(
          (m) => m.id !== "welcome" && !remoteIds.has(m.id),
        );
        return dedupeById([
          ...order.map((id) => byId.get(id)!).filter(Boolean),
          ...extras,
        ]);
      });
    },
    [],
  );

  useEffect(() => {
    if (!enabled) {
      setLiveChatOpen(false);
      return;
    }
    setLiveChatOpen(true);
    return () => setLiveChatOpen(false);
  }, [enabled]);

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
          rememberIds([WELCOME]);
        } else {
          const mapped = dedupeById(history.messages.map(mapHistoryMessage));
          setMessages(mapped);
          rememberIds(mapped);
          const unread = Math.max(
            history.unread_count ?? 0,
            unreadOnOpenRef.current,
          );
          const dividerId = firstUnreadDonnaId(mapped, unread);
          if (dividerId) {
            dividerSetRef.current = true;
            setNewMessageBeforeId(dividerId);
          }
        }
        historyReadyRef.current = true;
        refreshSummary();
      } catch {
        if (!cancelled) {
          setMessages([WELCOME]);
          rememberIds([WELCOME]);
          historyReadyRef.current = true;
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
  }, [enabled, refreshSummary, rememberIds]);

  useEffect(() => {
    if (!enabled || loadingHistory) {
      return;
    }
    let cancelled = false;

    const poll = async () => {
      try {
        const history = await fetchChatMessages(false);
        if (cancelled) return;
        if (history.messages.length === 0) return;
        mergeRemoteMessages(history.messages.map(mapHistoryMessage));
        refreshSummary();
      } catch {
        // Keep the open thread usable offline / on blips.
      }
    };

    const id = window.setInterval(() => {
      void poll();
    }, POLL_MS);

    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, [enabled, loadingHistory, mergeRemoteMessages, refreshSummary]);

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
    knownIdsRef.current.add(clientId);
    setMessages((prev) => {
      const withoutWelcome = prev.filter((m) => m.id !== "welcome");
      return [...withoutWelcome, userMsg];
    });
    playChatSendSound();
    setDraft("");
    setSending(true);
    const typingStartedAt = Date.now();
    try {
      const result = await sendChatCommand(text, clientId);
      const remaining = Math.max(0, TYPING_MIN_MS - (Date.now() - typingStartedAt));
      if (remaining > 0) {
        await wait(remaining);
      }
      setSending(false);
      await wait(TYPING_EXIT_MS);
      const replyId = result.reply_message_public_id || newId();
      if (result.user_message_public_id) {
        knownIdsRef.current.add(result.user_message_public_id);
        knownIdsRef.current.delete(clientId);
      }
      knownIdsRef.current.add(replyId);
      setMessages((prev) => {
        const serverUserId = result.user_message_public_id;
        // Poll may have already inserted the server user/reply ids while we
        // still held the optimistic clientId — drop or remap without dupes.
        const next = prev.flatMap((m) => {
          if (m.id !== clientId) {
            return [m];
          }
          if (!serverUserId) {
            return [m];
          }
          if (prev.some((other) => other.id === serverUserId)) {
            return [];
          }
          return [{ ...m, id: serverUserId }];
        });
        const withReply = next.some((m) => m.id === replyId)
          ? next
          : [
              ...next,
              {
                id: replyId,
                role: "donna" as const,
                text: result.reply,
                createdAt: Date.now(),
              },
            ];
        return dedupeById(withReply);
      });
      playChatReceiveSound();
    } catch {
      const remaining = Math.max(0, TYPING_MIN_MS - (Date.now() - typingStartedAt));
      if (remaining > 0) {
        await wait(remaining);
      }
      setSending(false);
      await wait(TYPING_EXIT_MS);
      const errId = newId();
      knownIdsRef.current.add(errId);
      setMessages((prev) => [
        ...prev,
        {
          id: errId,
          role: "donna",
          text: "Something went wrong on my end. Try again in a moment.",
          createdAt: Date.now(),
        },
      ]);
      playChatReceiveSound();
    } finally {
      setSending(false);
      refreshSummary();
    }
  }, [draft, sending, loadingHistory, refreshSummary]);

  return {
    messages,
    draft,
    setDraft,
    sending,
    loadingHistory,
    newMessageBeforeId,
    send,
    bottomRef,
  };
}
