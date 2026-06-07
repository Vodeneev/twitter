export interface User {
  id: string;
  email: string;
  username: string;
  displayName: string;
  bio: string;
  location: string;
  website: string;
  avatarUrl: string;
  headerUrl: string;
  isAdmin: boolean;
  isBanned: boolean;
  followersCount: number;
  followingCount: number;
  yapsCount: number;
  emailVerified: boolean;
  createdAt: string;
  following: boolean;
}

export interface Author {
  id: string;
  username: string;
  displayName: string;
  avatarUrl: string;
}

export interface Media {
  id: string;
  url: string;
  position: number;
}

export interface Yap {
  id: string;
  author: Author;
  content: string;
  replyToId?: string;
  quoteOfId?: string;
  quoteOf?: Yap;
  media: Media[];
  likesCount: number;
  repostsCount: number;
  repliesCount: number;
  quotesCount: number;
  bookmarksCount: number;
  liked: boolean;
  reposted: boolean;
  bookmarked: boolean;
  repostedBy?: Author;
  repostedAt?: string;
  createdAt: string;
}

export interface Page<T> {
  items: T[];
  nextCursor?: string | null;
}

export type NotificationType = 'like' | 'follow' | 'reply' | 'repost' | 'quote' | 'mention';

export interface AppNotification {
  id: string;
  type: NotificationType;
  actor: Author;
  yapId?: string;
  yapPreview?: string;
  read: boolean;
  createdAt: string;
}

export interface Message {
  id: string;
  conversationId: string;
  senderId: string;
  body: string;
  createdAt: string;
}

export interface Conversation {
  id: string;
  other: Author;
  lastMessage?: Message;
  unread: number;
  lastMessageAt: string;
}
