"use client";

import { useQuery, useQueryClient } from "@tanstack/react-query";

import { fetchChatSummary } from "./Chat.api";
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

export function useDonnaThreadSummary() {
  const queryClient = useQueryClient();
  const query = useQuery({
    queryKey: chatSummaryQueryKey,
    queryFn: fetchChatSummary,
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
