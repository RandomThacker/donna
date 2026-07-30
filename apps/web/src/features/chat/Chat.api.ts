import { apiRequest } from "@/lib/api/client";

import type {
  ChatCommandResponse,
  ChatHistoryResponse,
  ChatSummaryResponse,
} from "./Chat.types";

export async function sendChatCommand(
  message: string,
  clientMessageId?: string,
): Promise<ChatCommandResponse> {
  return apiRequest<ChatCommandResponse>("/api/v1/chat/command", {
    method: "POST",
    body: {
      message,
      ...(clientMessageId ? { client_message_id: clientMessageId } : {}),
    },
  });
}

export async function fetchChatMessages(): Promise<ChatHistoryResponse> {
  return apiRequest<ChatHistoryResponse>("/api/v1/chat/messages");
}

export async function fetchChatSummary(): Promise<ChatSummaryResponse> {
  return apiRequest<ChatSummaryResponse>("/api/v1/chat/summary");
}
