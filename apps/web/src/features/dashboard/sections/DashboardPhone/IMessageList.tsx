import { Icon } from "@/components/common";

import { iMessageStyles as styles } from "./DashboardPhone.styles";
import type { IMessageListProps } from "./DashboardPhone.types";

export function IMessageList({ conversations, onOpen, onClose }: IMessageListProps) {
  return (
    <div className={styles.listRoot}>
      <div className={styles.listHeader}>
        <div className={styles.listTopRow}>
          <button type="button" className={styles.edit}>
            Edit
          </button>
          {onClose ? (
            <button
              type="button"
              className={styles.compose}
              aria-label="Close"
              onClick={onClose}
            >
              <Icon name="close" className="h-5 w-5" />
            </button>
          ) : (
            <button type="button" className={styles.compose} aria-label="Compose">
              <Icon name="compose" className="h-5 w-5" />
            </button>
          )}
        </div>
        <h2 className={styles.listTitle}>Messages</h2>
        <div className={styles.search}>
          <Icon name="search" className="h-[15px] w-[15px]" />
          <span>Search</span>
        </div>
      </div>

      <div className={styles.listBody}>
        {conversations.map((conversation) => (
          <button
            key={conversation.id}
            type="button"
            className={styles.row}
            onClick={() => onOpen(conversation.id)}
          >
            <span className={styles.avatar}>
              {conversation.name.slice(0, 1)}
            </span>
            <span className={styles.rowMain}>
              <span className={styles.rowTop}>
                <span className={styles.rowName}>{conversation.name}</span>
                <span className={styles.rowMeta}>
                  <span className={styles.rowTime}>{conversation.time}</span>
                  <Icon name="chevronRight" className={styles.chevron} />
                </span>
              </span>
              <span className={styles.rowBottom}>
                <span className={styles.rowPreview}>{conversation.preview}</span>
                {conversation.unread > 0 ? (
                  <span className={styles.unread}>{conversation.unread}</span>
                ) : null}
              </span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
