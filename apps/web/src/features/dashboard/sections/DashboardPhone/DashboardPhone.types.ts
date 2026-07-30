import type { DashboardPhone, IMessageConversation } from "../../Dashboard.types";

export type DashboardPhoneProps = {
  phone: DashboardPhone;
};

export type IMessageListProps = {
  conversations: IMessageConversation[];
  onOpen: (id: string) => void;
  onClose?: () => void;
};

export type IMessageChatProps = {
  conversation: IMessageConversation;
  onBack: () => void;
  onClose?: () => void;
  /** Wire composer to command chat API (Donna thread). */
  live?: boolean;
  initialDraft?: string;
  showBack?: boolean;
};

export type DashboardPhoneFullscreenProps = {
  phone: DashboardPhone;
  onClose: () => void;
  exiting?: boolean;
  onCloseComplete?: () => void;
};
