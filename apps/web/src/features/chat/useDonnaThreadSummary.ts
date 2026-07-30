"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";

import { fetchChatMessages, fetchChatSummary } from "./Chat.api";
import type { ChatSummaryResponse } from "./Chat.types";

export const chatSummaryQueryKey = ["chat", "summary"] as const;

const EMPTY_PREVIEW = "Tell Donna what to do…";

function formatListTime(iso?: string | null): string {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat(undefined, {
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

function previewFromMessages(
  messages: { content: string; created_at: string }[],
): Pick<ChatSummaryResponse, "preview" | "last_message_at" | "unread_count"> {
  const last = messages[messages.length - 1];
  if (!last) {
    return { preview: "", last_message_at: null, unread_count: 0 };
  }
  return {
    preview: last.content,
    last_message_at: last.created_at,
    unread_count: 0,
  };
}

async function loadDonnaThreadSummary(): Promise<ChatSummaryResponse> {
  try {
    const summary = await fetchChatSummary();
    if (summary.preview?.trim()) {
      return summary;
    }
    // Summary exists but empty — fall through to messages for preview.
  } catch {
    // /chat/summary may be missing on older API deploys.
  }

  const history = await fetchChatMessages(false);
  const derived = previewFromMessages(history.messages);
  return {
    conversation_id: history.conversation_id,
    conversation_public_id: history.conversation_public_id,
    unread_count: history.unread_count ?? derived.unread_count,
    preview: derived.preview,
    last_message_at: derived.last_message_at,
  };
}

export function useDonnaThreadSummary() {
  const queryClient = useQueryClient();
  const query = useQuery({
    queryKey: chatSummaryQueryKey,
    queryFn: loadDonnaThreadSummary,
    refetchInterval: 30_000,
    staleTime: 10_000,
  });

  const data: ChatSummaryResponse | undefined = query.data;
  const preview = data?.preview?.trim() || EMPTY_PREVIEW;
  const unread = data?.unread_count ?? 0;
  const time = formatListTime(data?.last_message_at);

  return {
    preview,
    time,
    unread,
    isLoading: query.isLoading,
    refresh: () =>
      queryClient.invalidateQueries({ queryKey: chatSummaryQueryKey }),
  };
}
