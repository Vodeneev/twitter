import type {
  AppNotification,
  Conversation,
  Message,
  Page,
  User,
  Yap,
} from './types';

export class ApiError extends Error {
  status: number;
  code: string;
  field?: string;
  constructor(status: number, code: string, message: string, field?: string) {
    super(message);
    this.status = status;
    this.code = code;
    this.field = field;
  }
}

function apiBase(): string {
  return process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080';
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase()}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...(init?.headers ?? {}),
    },
  });
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    const err = data?.error;
    throw new ApiError(
      res.status,
      err?.code ?? 'internal_error',
      err?.message ?? 'request failed',
      err?.field,
    );
  }
  return data as T;
}

function qs(params: Record<string, string | number | undefined | null>): string {
  const sp = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === '') continue;
    sp.set(k, String(v));
  }
  const s = sp.toString();
  return s ? `?${s}` : '';
}

export interface RegisterInput {
  username: string;
  email: string;
  displayName: string;
  password: string;
  locale?: string;
}

export interface LoginInput {
  identifier: string;
  password: string;
  rememberMe: boolean;
}

export const api = {
  base: apiBase,

  // --- auth ---
  me: () => request<{ user: User | null }>('/api/auth/me'),
  register: (input: RegisterInput) =>
    request<{ user: User; verificationEmailSent: boolean }>('/api/auth/register', {
      method: 'POST',
      body: JSON.stringify(input),
    }),
  login: (input: LoginInput) =>
    request<{ user: User }>('/api/auth/login', { method: 'POST', body: JSON.stringify(input) }),
  logout: () => request<void>('/api/auth/logout', { method: 'POST', body: JSON.stringify({}) }),
  verifyEmail: (token: string) =>
    request<{ user: User }>('/api/auth/verify-email', { method: 'POST', body: JSON.stringify({ token }) }),
  resendVerification: (email: string, locale?: string) =>
    request<void>('/api/auth/resend-verification', { method: 'POST', body: JSON.stringify({ email, locale }) }),
  forgotPassword: (email: string, locale?: string) =>
    request<void>('/api/auth/forgot-password', { method: 'POST', body: JSON.stringify({ email, locale }) }),
  resetPassword: (token: string, password: string) =>
    request<void>('/api/auth/reset-password', { method: 'POST', body: JSON.stringify({ token, password }) }),

  // --- profile ---
  updateMe: (input: Partial<Pick<User, 'displayName' | 'bio' | 'location' | 'website'>> & { avatarKey?: string; headerKey?: string }) =>
    request<{ user: User }>('/api/me', { method: 'PATCH', body: JSON.stringify(input) }),
  getProfile: (username: string) => request<{ user: User }>(`/api/users/${encodeURIComponent(username)}`),
  followers: (username: string) => request<{ items: User[] }>(`/api/users/${encodeURIComponent(username)}/followers`),
  followingList: (username: string) => request<{ items: User[] }>(`/api/users/${encodeURIComponent(username)}/following`),
  suggestions: (limit = 5) => request<{ items: User[] }>(`/api/me/suggestions${qs({ limit })}`),
  follow: (username: string) => request<{ following: boolean }>(`/api/users/${encodeURIComponent(username)}/follow`, { method: 'PUT', body: JSON.stringify({}) }),
  unfollow: (username: string) => request<{ following: boolean }>(`/api/users/${encodeURIComponent(username)}/follow`, { method: 'DELETE' }),

  // --- timelines ---
  homeTimeline: (cursor?: string) => request<Page<Yap>>(`/api/timeline/home${qs({ cursor })}`),
  globalTimeline: (cursor?: string) => request<Page<Yap>>(`/api/timeline/global${qs({ cursor })}`),
  userYaps: (username: string, cursor?: string) => request<Page<Yap>>(`/api/users/${encodeURIComponent(username)}/yaps${qs({ cursor })}`),
  userReplies: (username: string, cursor?: string) => request<Page<Yap>>(`/api/users/${encodeURIComponent(username)}/replies${qs({ cursor })}`),
  userMedia: (username: string, cursor?: string) => request<Page<Yap>>(`/api/users/${encodeURIComponent(username)}/media${qs({ cursor })}`),
  userLikes: (username: string, cursor?: string) => request<Page<Yap>>(`/api/users/${encodeURIComponent(username)}/likes${qs({ cursor })}`),
  bookmarks: (cursor?: string) => request<Page<Yap>>(`/api/bookmarks${qs({ cursor })}`),
  hashtag: (tag: string, cursor?: string) => request<Page<Yap>>(`/api/hashtags/${encodeURIComponent(tag)}${qs({ cursor })}`),

  // --- yaps ---
  getYap: (id: string) => request<{ yap: Yap }>(`/api/yaps/${id}`),
  thread: (id: string) => request<{ yap: Yap; ancestors: Yap[] }>(`/api/yaps/${id}/thread`),
  replies: (id: string, cursor?: string) => request<Page<Yap>>(`/api/yaps/${id}/replies${qs({ cursor })}`),
  createYap: (input: { content: string; replyToId?: string; quoteOfId?: string; mediaKeys?: string[] }) =>
    request<{ yap: Yap }>('/api/yaps', { method: 'POST', body: JSON.stringify(input) }),
  deleteYap: (id: string) => request<void>(`/api/yaps/${id}`, { method: 'DELETE' }),
  like: (id: string) => request<void>(`/api/yaps/${id}/like`, { method: 'PUT', body: JSON.stringify({}) }),
  unlike: (id: string) => request<void>(`/api/yaps/${id}/like`, { method: 'DELETE' }),
  repost: (id: string) => request<void>(`/api/yaps/${id}/repost`, { method: 'PUT', body: JSON.stringify({}) }),
  unrepost: (id: string) => request<void>(`/api/yaps/${id}/repost`, { method: 'DELETE' }),
  bookmark: (id: string) => request<void>(`/api/yaps/${id}/bookmark`, { method: 'PUT', body: JSON.stringify({}) }),
  unbookmark: (id: string) => request<void>(`/api/yaps/${id}/bookmark`, { method: 'DELETE' }),

  // --- search ---
  searchUsers: (q: string) => request<{ items: User[] }>(`/api/search/users${qs({ q })}`),
  searchYaps: (q: string, cursor?: string) => request<Page<Yap>>(`/api/search/yaps${qs({ q, cursor })}`),

  // --- notifications ---
  notifications: (cursor?: string) => request<Page<AppNotification>>(`/api/notifications${qs({ cursor })}`),
  unreadCount: () => request<{ count: number }>('/api/notifications/unread-count'),
  markNotificationsRead: () => request<void>('/api/notifications/read', { method: 'POST', body: JSON.stringify({}) }),

  // --- DM ---
  conversations: () => request<{ items: Conversation[] }>('/api/conversations'),
  openConversation: (username: string) =>
    request<{ conversationId: string; other: User }>('/api/conversations', { method: 'POST', body: JSON.stringify({ username }) }),
  messages: (convId: string, cursor?: string) =>
    request<{ items: Message[]; nextCursor?: string | null }>(`/api/conversations/${convId}/messages${qs({ cursor })}`),
  sendMessage: (convId: string, body: string) =>
    request<{ message: Message }>(`/api/conversations/${convId}/messages`, { method: 'POST', body: JSON.stringify({ body }) }),
  markConversationRead: (convId: string) =>
    request<void>(`/api/conversations/${convId}/read`, { method: 'POST', body: JSON.stringify({}) }),

  // --- media ---
  presign: (contentType: string, kind: 'avatar' | 'header' | 'yap') =>
    request<{ uploadUrl: string; key: string; publicUrl: string }>('/api/media/presign', {
      method: 'POST',
      body: JSON.stringify({ contentType, kind }),
    }),
};

export async function uploadToStorage(uploadUrl: string, file: File): Promise<void> {
  const res = await fetch(uploadUrl, {
    method: 'PUT',
    headers: { 'Content-Type': file.type },
    body: file,
  });
  if (!res.ok) {
    throw new Error(`upload failed: HTTP ${res.status}`);
  }
}
