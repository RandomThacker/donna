/**
 * Chat send / receive / notify sounds from src/lib/sound.
 * Also raises a browser Notification when the tab is in the background.
 */

import notifySrc from "@/lib/sound/iphone_sms_tone.mp3";
import receiveSrc from "@/lib/sound/iphone_receive_sms.mp3";
import sentSrc from "@/lib/sound/iphone_sent_sms.mp3";

const SOUND = {
  sent: sentSrc,
  receive: receiveSrc,
  notify: notifySrc,
} as const;

let unlockBound = false;
let liveChatOpen = false;
let lastReceiveAt = 0;
const players = new Map<string, HTMLAudioElement>();

function getPlayer(src: string): HTMLAudioElement | null {
  if (typeof window === "undefined") return null;
  let audio = players.get(src);
  if (!audio) {
    audio = new Audio(src);
    audio.preload = "auto";
    players.set(src, audio);
  }
  return audio;
}

function playSrc(src: string): void {
  bindUnlock();
  try {
    const audio = getPlayer(src);
    if (!audio) return;
    audio.currentTime = 0;
    void audio.play().catch(() => {
      // Autoplay blocked until a gesture — ignore.
    });
  } catch {
    // Audio unavailable — ignore.
  }
}

function bindUnlock(): void {
  if (typeof window === "undefined" || unlockBound) return;
  unlockBound = true;
  const unlock = () => {
    for (const src of Object.values(SOUND)) {
      const audio = getPlayer(src);
      if (!audio) continue;
      // Prime decode + satisfy gesture-gated playback.
      audio
        .play()
        .then(() => {
          audio.pause();
          audio.currentTime = 0;
        })
        .catch(() => {});
    }
    void ensureBrowserNotificationPermission();
  };
  window.addEventListener("pointerdown", unlock, { once: true, passive: true });
  window.addEventListener("keydown", unlock, { once: true });
}

export async function ensureBrowserNotificationPermission(): Promise<
  NotificationPermission | "unsupported"
> {
  if (typeof window === "undefined" || !("Notification" in window)) {
    return "unsupported";
  }
  if (
    Notification.permission === "granted" ||
    Notification.permission === "denied"
  ) {
    return Notification.permission;
  }
  try {
    return await Notification.requestPermission();
  } catch {
    return Notification.permission;
  }
}

function showBrowserNotification(body: string): void {
  if (typeof window === "undefined" || !("Notification" in window)) return;
  if (Notification.permission !== "granted") return;
  // Foreground + chat open: in-app sound is enough.
  if (!document.hidden && liveChatOpen) return;

  try {
    const note = new Notification("Donna", {
      body: body.trim() || "New message",
      icon: "/icons/donna-icon.svg",
      tag: "donna-chat-message",
      renotify: true,
      silent: false,
    });
    note.onclick = () => {
      window.focus();
      note.close();
    };
  } catch {
    // Notification constructor can throw on some platforms — ignore.
  }
}

/** Sent-message tone. */
export function playChatSendSound(): void {
  playSrc(SOUND.sent);
}

/** Received-message tone (while thread is open). */
export function playChatReceiveSound(): void {
  lastReceiveAt = Date.now();
  playSrc(SOUND.receive);
}

/**
 * Unread / background alert: plays notify tone and, when useful,
 * raises a free Browser Notification (no push API key required).
 */
export function playChatNotificationSound(preview = ""): void {
  if (Date.now() - lastReceiveAt < 1200) {
    return;
  }
  playSrc(SOUND.notify);
  // Always try browser notification when the page isn’t focused.
  if (typeof document !== "undefined" && document.hidden) {
    showBrowserNotification(preview);
  } else if (!liveChatOpen) {
    showBrowserNotification(preview);
  }
}

export function setLiveChatOpen(open: boolean): void {
  liveChatOpen = open;
}

export function isLiveChatOpen(): boolean {
  return liveChatOpen;
}

bindUnlock();
