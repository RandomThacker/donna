import { apiRequest } from "@/lib/api/client";

import type { ChatCommandResponse } from "./Chat.types";

export async function sendChatCommand(
  message: string,
): Promise<ChatCommandResponse> {
  return apiRequest<ChatCommandResponse>("/api/v1/chat/command", {
    method: "POST",
    body: { message },
  });
}
