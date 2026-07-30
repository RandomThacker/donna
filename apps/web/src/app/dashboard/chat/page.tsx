"use client";

import { RequireAuth } from "@/features/auth";
import { Chat } from "@/features/chat";

export default function ChatPage() {
  return (
    <RequireAuth>
      <Chat />
    </RequireAuth>
  );
}
