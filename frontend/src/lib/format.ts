export function timeAgo(iso: string, locale: string): string {
  const then = new Date(iso).getTime();
  const now = Date.now();
  const sec = Math.max(0, Math.floor((now - then) / 1000));
  const rtf = new Intl.RelativeTimeFormat(locale, { numeric: 'auto', style: 'narrow' });
  if (sec < 60) return rtf.format(-sec, 'second');
  const min = Math.floor(sec / 60);
  if (min < 60) return rtf.format(-min, 'minute');
  const hr = Math.floor(min / 60);
  if (hr < 24) return rtf.format(-hr, 'hour');
  const day = Math.floor(hr / 24);
  if (day < 7) return rtf.format(-day, 'day');
  return new Date(iso).toLocaleDateString(locale, { day: 'numeric', month: 'short' });
}

export function formatDate(iso: string, locale: string): string {
  return new Date(iso).toLocaleDateString(locale, { day: 'numeric', month: 'long', year: 'numeric' });
}

/** Clock time for a message, e.g. 14:35. */
export function formatMessageTime(iso: string, locale: string): string {
  return new Date(iso).toLocaleTimeString(locale, { hour: '2-digit', minute: '2-digit' });
}

/** Time if today, otherwise a short date — for conversation lists. */
export function formatConversationTime(iso: string, locale: string): string {
  const d = new Date(iso);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return formatMessageTime(iso, locale);
  }
  return d.toLocaleDateString(locale, { day: 'numeric', month: 'short' });
}

export type Segment =
  | { type: 'text'; value: string }
  | { type: 'hashtag'; value: string }
  | { type: 'mention'; value: string }
  | { type: 'url'; value: string };

const tokenRe = /(#[\p{L}\p{N}_]{1,50})|(@[a-zA-Z0-9_]{1,20})|(https?:\/\/[^\s]+)/gu;

// tokenize splits yap content into renderable segments (links, hashtags, mentions).
export function tokenize(content: string): Segment[] {
  const out: Segment[] = [];
  let last = 0;
  for (const m of content.matchAll(tokenRe)) {
    const idx = m.index ?? 0;
    if (idx > last) out.push({ type: 'text', value: content.slice(last, idx) });
    const tok = m[0];
    if (tok.startsWith('#')) out.push({ type: 'hashtag', value: tok });
    else if (tok.startsWith('@')) out.push({ type: 'mention', value: tok });
    else out.push({ type: 'url', value: tok });
    last = idx + tok.length;
  }
  if (last < content.length) out.push({ type: 'text', value: content.slice(last) });
  return out;
}

export function initialsAvatar(name: string): string {
  const trimmed = name.trim();
  if (!trimmed) return '?';
  return trimmed.charAt(0).toUpperCase();
}
