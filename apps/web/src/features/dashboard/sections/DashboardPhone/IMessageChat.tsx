import { Icon } from "@/components/common";
import { cn } from "@/lib/cn";

import type { IMessageBubble } from "../../Dashboard.types";
import { BubbleTail } from "./BubbleTail";
import { iMessageStyles as styles } from "./DashboardPhone.styles";
import type { IMessageChatProps } from "./DashboardPhone.types";

function isLastInGroup(messages: IMessageBubble[], index: number) {
  const current = messages[index];
  const next = messages[index + 1];
  if (!current) return true;
  return !next || next.role !== current.role;
}

function isFirstInGroup(messages: IMessageBubble[], index: number) {
  const current = messages[index];
  const prev = messages[index - 1];
  if (!current) return true;
  return !prev || prev.role !== current.role;
}

export function IMessageChat({ conversation, onBack }: IMessageChatProps) {
  return (
    <div className={styles.chatRoot}>
      <header className={styles.chatNav}>
        <button type="button" className={styles.back} onClick={onBack}>
          <Icon name="chevronLeft" className="h-4 w-4" />
          <span>Messages</span>
        </button>
        <div className={styles.chatTitleWrap}>
          <span className={styles.chatAvatar}>{conversation.name.slice(0, 1)}</span>
          <span className={styles.chatName}>{conversation.name}</span>
        </div>
        <button type="button" className={styles.chatInfo} aria-label="Info">
          <Icon name="info" className="h-4 w-4" />
        </button>
      </header>

      <div className={styles.chatBody}>
        <p className={styles.stamp}>Today {conversation.messages[0]?.time}</p>
        {conversation.messages.map((message, index) => {
          const incoming = message.role === "donna";
          const last = isLastInGroup(conversation.messages, index);
          const first = isFirstInGroup(conversation.messages, index);

          return (
            <div
              key={message.id}
              className={cn(
                styles.bubbleWrap,
                incoming ? styles.bubbleWrapIn : styles.bubbleWrapOut,
                first ? styles.bubbleSpaced : styles.bubbleGrouped,
              )}
            >
              <div
                className={cn(
                  styles.bubble,
                  incoming ? styles.bubbleIn : styles.bubbleOut,
                  last
                    ? incoming
                      ? styles.bubbleInLast
                      : styles.bubbleOutLast
                    : incoming
                      ? styles.bubbleInMiddle
                      : styles.bubbleOutMiddle,
                )}
              >
                {message.text}
              </div>
              {last ? (
                <BubbleTail
                  side={incoming ? "in" : "out"}
                  className={incoming ? styles.tailIn : styles.tailOut}
                />
              ) : null}
            </div>
          );
        })}
      </div>

      <div className={styles.composer}>
        <button type="button" className={styles.plus} aria-label="Apps">
          <Icon name="plus" className="h-4 w-4" />
        </button>
        <div className={styles.inputShell}>
          <input
            type="text"
            className={styles.input}
            placeholder="iMessage"
            aria-label="iMessage"
            readOnly
          />
          <button type="button" className={styles.mic} aria-label="Dictation">
            <Icon name="mic" className="h-[18px] w-[18px]" />
          </button>
        </div>
      </div>
    </div>
  );
}
