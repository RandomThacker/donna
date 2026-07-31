/**
 * Chat send / receive / notify sounds.
 *
 * Uses Web Audio (one AudioContext + decoded buffers) so:
 * - Unlock on a user gesture covers later async receive plays (mobile).
 * - Send no longer races with a silent "prime" on the same HTMLAudio element
 *   (that race killed send on desktop and receive on mobile).
 *
 * Background tabs: page audio is often suspended; we also raise a SW
 * notification so the OS can alert.
 */

import notifyBundled from "@/lib/sound/iphone_sms_tone.mp3";
import receiveBundled from "@/lib/sound/iphone_receive_sms.mp3";
import sentBundled from "@/lib/sound/iphone_sent_sms.mp3";

type SoundKey = "sent" | "receive" | "notify";

/** Stable public URLs (best for SW notifications + caching). */
const SOUND: Record<SoundKey, string> = {
  sent: "/sounds/iphone_sent_sms.mp3",
  receive: "/sounds/iphone_receive_sms.mp3",
  notify: "/sounds/iphone_sms_tone.mp3",
};

/** Bundled fallbacks if public files are missing in some deploys. */
const SOUND_FALLBACK: Record<SoundKey, string> = {
  sent: sentBundled,
  receive: receiveBundled,
  notify: notifyBundled,
};

let unlockBound = false;
let audioUnlocked = false;
let liveChatOpen = false;
let lastReceiveAt = 0;
let ctx: AudioContext | null = null;
const buffers = new Map<SoundKey, AudioBuffer>();
const loading = new Map<SoundKey, Promise<AudioBuffer | null>>();

function getAudioContext(): AudioContext | null {
  if (typeof window === "undefined") return null;
  if (ctx) return ctx;
  const AC =
    window.AudioContext ||
    (window as unknown as { webkitAudioContext?: typeof AudioContext })
      .webkitAudioContext;
  if (!AC) return null;
  ctx = new AC();
  return ctx;
}

async function loadBuffer(key: SoundKey): Promise<AudioBuffer | null> {
  const cached = buffers.get(key);
  if (cached) return cached;

  const inflight = loading.get(key);
  if (inflight) return inflight;

  const task = (async () => {
    const audioCtx = getAudioContext();
    if (!audioCtx) return null;

    for (const url of [SOUND[key], SOUND_FALLBACK[key]]) {
      try {
        const res = await fetch(url);
        if (!res.ok) continue;
        const raw = await res.arrayBuffer();
        const buffer = await audioCtx.decodeAudioData(raw.slice(0));
        buffers.set(key, buffer);
        return buffer;
      } catch {
        // Try next URL.
      }
    }
    return null;
  })().finally(() => {
    loading.delete(key);
  });

  loading.set(key, task);
  return task;
}

async function resumeContext(): Promise<boolean> {
  const audioCtx = getAudioContext();
  if (!audioCtx) return false;
  try {
    if (audioCtx.state === "suspended") {
      await audioCtx.resume();
    }
    audioUnlocked = audioCtx.state === "running";
    return audioUnlocked;
  } catch {
    return false;
  }
}

async function playKey(key: SoundKey): Promise<boolean> {
  const audioCtx = getAudioContext();
  if (!audioCtx) return false;

  await resumeContext();
  const buffer = await loadBuffer(key);
  if (!buffer) return false;

  // Context can suspend again while decoding on some mobile browsers.
  await resumeContext();
  if (audioCtx.state !== "running") return false;

  try {
    const source = audioCtx.createBufferSource();
    source.buffer = buffer;
    const gain = audioCtx.createGain();
    gain.gain.value = 1;
    source.connect(gain);
    gain.connect(audioCtx.destination);
    source.start(0);
    return true;
  } catch {
    return false;
  }
}

function playSound(key: SoundKey): void {
  bindUnlock();
  void playKey(key);
}

/**
 * Call from a user-gesture handler (send tap, fab tap) so later async
 * receive/notify plays are allowed on mobile.
 */
export function primeChatAudio(): void {
  bindUnlock();
  // Resume must start inside the gesture stack — don't await first.
  const audioCtx = getAudioContext();
  if (audioCtx?.state === "suspended") {
    void audioCtx.resume().then(() => {
      audioUnlocked = audioCtx.state === "running";
    });
  } else if (audioCtx?.state === "running") {
    audioUnlocked = true;
  }
  void (async () => {
    if (!(await resumeContext())) return;
    await Promise.all(
      (Object.keys(SOUND) as SoundKey[]).map((key) => loadBuffer(key)),
    );
  })();
}

function bindUnlock(): void {
  if (typeof window === "undefined" || unlockBound) return;
  unlockBound = true;

  const onGesture = () => {
    const audioCtx = getAudioContext();
    if (audioCtx?.state === "suspended") {
      void audioCtx.resume().then(() => {
        audioUnlocked = audioCtx.state === "running";
      });
    } else if (audioCtx?.state === "running") {
      audioUnlocked = true;
    }
    void (async () => {
      await resumeContext();
      await Promise.all(
        (Object.keys(SOUND) as SoundKey[]).map((key) => loadBuffer(key)),
      );
      if (audioUnlocked) {
        window.removeEventListener("touchstart", onGesture);
        window.removeEventListener("pointerdown", onGesture);
        window.removeEventListener("click", onGesture);
      }
    })();
  };

  window.addEventListener("touchstart", onGesture, { passive: true });
  window.addEventListener("pointerdown", onGesture, { passive: true });
  window.addEventListener("click", onGesture);

  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible" && audioUnlocked) {
      void resumeContext();
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

/** Sent-message tone. Must stay inside the send gesture call stack start. */
export function playChatSendSound(): void {
  bindUnlock();
  // Kick resume synchronously in the gesture; then decode/play.
  const audioCtx = getAudioContext();
  if (audioCtx?.state === "suspended") {
    void audioCtx.resume().then(() => {
      audioUnlocked = audioCtx.state === "running";
    });
  } else if (audioCtx?.state === "running") {
    audioUnlocked = true;
  }
  void ensureBrowserNotificationPermission();
  void (async () => {
    await playKey("sent");
    void Promise.all([loadBuffer("receive"), loadBuffer("notify")]);
  })();
}

/** Received-message tone (while thread is open). */
export function playChatReceiveSound(): void {
  lastReceiveAt = Date.now();
  playSound("receive");
}

/**
 * Unread / background alert.
 * Plays in-app tone when possible; raises a SW notification when useful
 * so mobile PWAs get an OS sound.
 */
export function playChatNotificationSound(preview = ""): void {
  if (Date.now() - lastReceiveAt < 1200) {
    return;
  }

  const shouldNotify =
    typeof document !== "undefined" && (document.hidden || !liveChatOpen);

  if (!document.hidden) {
    playSound("notify");
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
