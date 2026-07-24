import type { DashboardPhone, IMessageConversation } from "../../Dashboard.types";

export type DashboardPhoneProps = {
  phone: DashboardPhone;
};

export type IMessageListProps = {
  conversations: IMessageConversation[];
  onOpen: (id: string) => void;
};

export type IMessageChatProps = {
  conversation: IMessageConversation;
  onBack: () => void;
};
