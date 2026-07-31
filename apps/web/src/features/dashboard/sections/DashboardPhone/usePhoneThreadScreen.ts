"use client";

import { useMemo, useState } from "react";

import { useDonnaThreadSummary } from "@/features/chat/useDonnaThreadSummary";

import type { IMessageConversation } from "../../Dashboard.types";

function withDonnaSummary(
  conversations: IMessageConversation[],
  preview: string,
  time: string,
  unread: number,
): IMessageConversation[] {
  return conversations.map((conversation) =>
    conversation.id === "donna"
      ? {
          ...conversation,
          preview,
          time: time || conversation.time,
          unread,
        }
      : conversation,
  );
}

/** Unread → chat list. Inbox clear → open Donna thread. */
export function initialPhoneThreadId(unread: number): string | null {
  return unread > 0 ? null : "donna";
}

export function usePhoneThreadScreen(conversations: IMessageConversation[]) {
  const donna = useDonnaThreadSummary();
  const [activeId, setActiveId] = useState<string | null>(null);
  const [seeded, setSeeded] = useState(false);

  // Seed once summary is ready so desktop/mobile open on the right screen.
  if (!seeded && !donna.isLoading) {
    setSeeded(true);
    setActiveId(initialPhoneThreadId(donna.unread));
  }

  const threads = useMemo(
    () =>
      withDonnaSummary(
        conversations,
        donna.preview,
        donna.time,
        donna.unread,
      ),
    [conversations, donna.preview, donna.time, donna.unread],
  );

  const activeConversation = useMemo(
    () => threads.find((item) => item.id === activeId) ?? null,
    [activeId, threads],
  );

  return {
    donna,
    threads,
    activeId,
    setActiveId,
    activeConversation,
  };
}
