/**
 * Chat send / receive / notify sounds.
 *
 * Mobile PWAs are strict: audio must be primed by a user gesture, and
 * background tabs usually cannot play HTMLAudio. For background alerts we
 * use the service worker Notification API so the OS plays a sound.
 */

import notifyBundled from "@/lib/sound/iphone_sms_tone.mp3";
import receiveBundled from "@/lib/sound/iphone_receive_sms.mp3";
import sentBundled from "@/lib/sound/iphone_sent_sms.mp3";

/** Stable public URLs (best for SW notifications + iOS). */
const SOUND = {
  sent: "/sounds/iphone_sent_sms.mp3",
  receive: "/sounds/iphone_receive_sms.mp3",
  notify: "/sounds/iphone_sms_tone.mp3",
} as const;

/** Bundled fallbacks if public files are missing in some deploys. */
const SOUND_FALLBACK = {
  sent: sentBundled,
  receive: receiveBundled,
  notify: notifyBundled,
} as const;

let unlockBound = false;
let audioUnlocked = false;
let liveChatOpen = false;
let lastReceiveAt = 0;
const players = new Map<string, HTMLAudioElement>();

function getPlayer(src: string): HTMLAudioElement | null {
  if (typeof window === "undefined") return null;
  let audio = players.get(src);
  if (!audio) {
    audio = new Audio(src);
    audio.preload = "auto";
    // Required for iOS / iPadOS — without this, play() is ignored.
    audio.setAttribute("playsinline", "true");
    audio.setAttribute("webkit-playsinline", "true");
    (audio as HTMLAudioElement & { playsInline?: boolean }).playsInline = true;
    audio.volume = 1;
    players.set(src, audio);
  }
  return audio;
}

async function tryPlay(src: string): Promise<boolean> {
  const audio = getPlayer(src);
  if (!audio) return false;
  try {
    audio.pause();
    audio.currentTime = 0;
    await audio.play();
    audioUnlocked = true;
    return true;
  } catch {
    // Try bundled URL once if public path failed to load.
    return false;
  }
}

function playSrc(src: string, fallback?: string): void {
  bindUnlock();
  void (async () => {
    if (await tryPlay(src)) return;
    if (fallback && fallback !== src) {
      await tryPlay(fallback);
    }
  })();
}

/**
 * Call from a user-gesture handler (send tap, nav tap) so later async
 * receive/notify plays are allowed on mobile.
 */
export function primeChatAudio(): void {
  bindUnlock();
  void unlockAudio();
}

async function unlockAudio(): Promise<void> {
  for (const key of Object.keys(SOUND) as Array<keyof typeof SOUND>) {
    const audio = getPlayer(SOUND[key]);
    if (!audio) continue;
    try {
      // Silent prime — volume 0 then restore. Still counts as a gesture unlock.
      const prev = audio.volume;
      audio.volume = 0.001;
      await audio.play();
      audio.pause();
      audio.currentTime = 0;
      audio.volume = prev || 1;
      audioUnlocked = true;
    } catch {
      // Keep trying on the next gesture.
    }
  }
  void ensureBrowserNotificationPermission();
}

function bindUnlock(): void {
  if (typeof window === "undefined" || unlockBound) return;
  unlockBound = true;

  const onGesture = () => {
    void unlockAudio().then(() => {
      if (audioUnlocked) {
        window.removeEventListener("touchstart", onGesture);
        window.removeEventListener("pointerdown", onGesture);
        window.removeEventListener("click", onGesture);
      }
    });
  };

  // Don't use { once: true } until unlock succeeds — first attempt often fails
  // on iOS if the gesture didn't reach audio in time.
  window.addEventListener("touchstart", onGesture, { passive: true });
  window.addEventListener("pointerdown", onGesture, { passive: true });
  window.addEventListener("click", onGesture);

  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") {
      // Re-warm players when returning to the PWA.
      if (audioUnlocked) {
        void unlockAudio();
      }
    }
  });
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

async function showBrowserNotification(body: string): Promise<void> {
  if (typeof window === "undefined" || !("Notification" in window)) return;
  if (Notification.permission !== "granted") return;
  if (!document.hidden && liveChatOpen) return;

  const title = "Donna";
  const options: NotificationOptions & { sound?: string; renotify?: boolean } = {
    body: body.trim() || "New message",
    icon: "/icons/icon-192.png",
    badge: "/icons/icon-192.png",
    tag: "donna-chat-message",
    silent: false,
    // Chrome Android may honor this; iOS uses the system sound instead.
    sound: SOUND.notify,
    data: { url: "/dashboard" },
  };

  // Service-worker notifications are what actually beep on mobile PWAs
  // when the app is backgrounded (page audio is suspended).
  try {
    if ("serviceWorker" in navigator) {
      const reg = await navigator.serviceWorker.ready;
      await reg.showNotification(title, options);
      return;
    }
  } catch {
    // Fall through to page Notification.
  }

  try {
    const note = new Notification(title, options);
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
  // Send is always inside a gesture — unlock + play together.
  primeChatAudio();
  playSrc(SOUND.sent, SOUND_FALLBACK.sent);
}

/** Received-message tone (while thread is open). */
export function playChatReceiveSound(): void {
  lastReceiveAt = Date.now();
  playSrc(SOUND.receive, SOUND_FALLBACK.receive);
}

/**
 * Unread / background alert.
 * Plays in-app tone when possible; always raises a SW notification when
 * useful so mobile PWAs get an OS sound.
 */
export function playChatNotificationSound(preview = ""): void {
  if (Date.now() - lastReceiveAt < 1200) {
    return;
  }

  const shouldNotify =
    typeof document !== "undefined" &&
    (document.hidden || !liveChatOpen);

  if (!document.hidden) {
    playSrc(SOUND.notify, SOUND_FALLBACK.notify);
  }

  if (shouldNotify) {
    void showBrowserNotification(preview);
  }
}

export function setLiveChatOpen(open: boolean): void {
  liveChatOpen = open;
}

export function isLiveChatOpen(): boolean {
  return liveChatOpen;
}

export function isChatAudioUnlocked(): boolean {
  return audioUnlocked;
}

bindUnlock();
