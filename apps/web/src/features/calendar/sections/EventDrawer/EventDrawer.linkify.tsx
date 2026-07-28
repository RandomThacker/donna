import type { ReactNode } from "react";

/**
 * Matches http(s) URLs, including Outlook/ICS style `<https://…>` wrappers.
 * Stops before whitespace and common delimiters that are not part of the URL.
 */
const URL_PATTERN =
  /<?(https?:\/\/[^\s<>"']+)>?/gi;

function cleanHref(raw: string): { href: string; trailing: string } {
  let href = raw;
  let trailing = "";
  // Peel sentence punctuation that often trails pasted links.
  while (/[.,;:!]+$/.test(href)) {
    trailing = href.slice(-1) + trailing;
    href = href.slice(0, -1);
  }
  return { href, trailing };
}

/** Turns plain description text into nodes with clickable http(s) links. */
export function linkifyPlainText(text: string): ReactNode[] {
  const nodes: ReactNode[] = [];
  let lastIndex = 0;
  let key = 0;
  const pattern = new RegExp(URL_PATTERN.source, URL_PATTERN.flags);

  for (const match of text.matchAll(pattern)) {
    const full = match[0];
    const rawUrl = match[1] ?? full;
    const index = match.index ?? 0;

    if (index > lastIndex) {
      nodes.push(text.slice(lastIndex, index));
    }

    const { href, trailing } = cleanHref(rawUrl);
    if (href) {
      nodes.push(
        <a
          key={`link-${key}`}
          href={href}
          target="_blank"
          rel="noopener noreferrer"
          className="break-all text-sky-400 underline-offset-2 hover:text-sky-300 hover:underline"
        >
          {href}
        </a>,
      );
      key += 1;
    }
    if (trailing) {
      nodes.push(trailing);
    }
    lastIndex = index + full.length;
  }

  if (lastIndex < text.length) {
    nodes.push(text.slice(lastIndex));
  }

  return nodes.length > 0 ? nodes : [text];
}
