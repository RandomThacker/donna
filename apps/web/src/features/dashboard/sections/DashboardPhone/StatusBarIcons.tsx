/** Minimal iOS-style status glyphs — color follows --phone-status. */
export function StatusBarIcons() {
  return (
    <span className="flex items-center gap-1.5 text-[var(--phone-status)]" aria-hidden>
      <svg width="16" height="10" viewBox="0 0 16 10" fill="currentColor">
        <rect x="0" y="6" width="2.5" height="4" rx="0.4" />
        <rect x="3.5" y="4" width="2.5" height="6" rx="0.4" />
        <rect x="7" y="2" width="2.5" height="8" rx="0.4" />
        <rect x="10.5" y="0" width="2.5" height="10" rx="0.4" />
      </svg>
      <svg width="14" height="10" viewBox="0 0 14 10" fill="currentColor">
        <path d="M7 2.2c1.7 0 3.2.7 4.3 1.8l-1.1 1.1A4.4 4.4 0 0 0 7 3.8c-1.2 0-2.3.5-3.1 1.3L2.8 4A5.8 5.8 0 0 1 7 2.2Zm0 2.8c.9 0 1.7.4 2.3 1L8.2 7A1.9 1.9 0 0 0 7 6.5c-.5 0-1 .2-1.3.6L4.6 6a3.3 3.3 0 0 1 2.4-1Zm0 2.8a1 1 0 1 1 0 2 1 1 0 0 1 0-2Z" />
      </svg>
      <svg width="22" height="10" viewBox="0 0 22 10" fill="none">
        <rect x="0.5" y="0.5" width="18" height="9" rx="2" stroke="currentColor" />
        <rect x="2" y="2" width="13" height="6" rx="1" fill="currentColor" />
        <path d="M20 3.2v3.6a1.4 1.4 0 0 0 0-3.6Z" fill="currentColor" />
      </svg>
    </span>
  );
}
