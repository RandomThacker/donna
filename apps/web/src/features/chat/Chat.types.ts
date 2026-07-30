export type ChatRole = "user" | "donna";

export type ChatMessage = {
  id: string;
  role: ChatRole;
  text: string;
  createdAt: number;
};

export type ChatCommandResponse = {
  reply: string;
  intent: string;
  conversation_public_id?: string;
  user_message_public_id?: string;
  reply_message_public_id?: string;
};

export type ChatHistoryMessage = {
  id: string;
  public_id: string;
  role: "user" | "assistant" | "system";
  content: string;
  created_at: string;
};

export type ChatHistoryResponse = {
  conversation_id: string;
  conversation_public_id: string;
  unread_count?: number;
  messages: ChatHistoryMessage[];
};

export type ChatSummaryResponse = {
  conversation_id: string;
  conversation_public_id: string;
  unread_count: number;
  preview: string;
  last_message_at?: string | null;
  last_message_role?: string;
};
