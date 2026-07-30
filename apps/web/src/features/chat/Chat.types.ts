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
};
