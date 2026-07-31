/**
 * Tiny store so the bottom bar can hide while the Donna phone sheet is open.
 */

type Listener = () => void;

let phoneOpen = false;
const listeners = new Set<Listener>();

export function setDonnaPhoneOpen(open: boolean): void {
  if (phoneOpen === open) return;
  phoneOpen = open;
  if (typeof document !== "undefined") {
    if (open) {
      document.documentElement.dataset.donnaPhoneOpen = "true";
    } else {
      delete document.documentElement.dataset.donnaPhoneOpen;
    }
  }
  for (const listener of listeners) {
    listener();
  }
}

export function getDonnaPhoneOpen(): boolean {
  return phoneOpen;
}

export function subscribeDonnaPhoneOpen(listener: Listener): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}
