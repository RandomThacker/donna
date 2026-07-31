"use client";

import { useDonnaThreadSummary } from "@/features/chat/useDonnaThreadSummary";
import { primeChatAudio } from "@/features/chat/chatSounds";
import { cn } from "@/lib/cn";

import { getDashboardContent } from "../../Dashboard.logic";
import { phoneFabStyles as styles } from "./DashboardPhone.styles";
import { DashboardPhoneFullscreen } from "./DashboardPhoneFullscreen";
import { useDonnaPhoneFab } from "./DonnaPhoneFab.logic";

export function DonnaPhoneFab() {
  const fab = useDonnaPhoneFab();
  const { data } = getDashboardContent();
  const donna = useDonnaThreadSummary();
  const hasNotification = donna.unread > 0;

  return (
    <>
      {!fab.open ? (
        <button
          type="button"
          className={cn(styles.button, fab.dragging && styles.buttonDragging)}
          style={fab.fabStyle}
          aria-label={
            hasNotification
              ? "Open Donna messages, new notifications"
              : "Open Donna messages"
          }
          onPointerDown={(event) => {
            primeChatAudio();
            fab.onPointerDown(event);
          }}
          onPointerMove={fab.onPointerMove}
          onPointerUp={fab.onPointerUp}
          onPointerCancel={fab.onPointerUp}
        >
          <span className={styles.mark} aria-hidden />
          {hasNotification ? (
            <span className={styles.badge} aria-hidden />
          ) : null}
        </button>
      ) : null}
      {fab.open ? (
        <DashboardPhoneFullscreen
          phone={data.phone}
          exiting={fab.exiting}
          onClose={fab.requestClose}
          onCloseComplete={fab.finishClose}
        />
      ) : null}
    </>
  );
}
