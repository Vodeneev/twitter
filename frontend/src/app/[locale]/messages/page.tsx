'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useSearchParams } from 'next/navigation';
import { useLocale, useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { RequireAuth } from '@/components/require-auth';
import { Avatar } from '@/components/avatar';
import { useSession } from '@/components/session-provider';
import { useRealtime, type RealtimeEvent } from '@/hooks/use-realtime';
import { api } from '@/lib/api';
import { formatConversationTime, formatMessageTime } from '@/lib/format';
import type { Author, Conversation, Message } from '@/lib/types';

export const dynamic = 'force-dynamic';

interface Active {
  id: string;
  other: Author;
}

function MessagesInner() {
  const t = useTranslations('messages');
  const locale = useLocale();
  const { user } = useSession();
  const params = useSearchParams();
  const initialConv = params.get('c');

  const [conversations, setConversations] = useState<Conversation[]>([]);
  const [active, setActive] = useState<Active | null>(null);
  const [messages, setMessages] = useState<Message[]>([]);
  const [body, setBody] = useState('');
  const endRef = useRef<HTMLDivElement>(null);

  const loadConversations = useCallback(async () => {
    const { items } = await api.conversations();
    setConversations(items ?? []);
    return items ?? [];
  }, []);

  useEffect(() => {
    loadConversations().then((items) => {
      if (initialConv) {
        const c = items.find((x) => x.id === initialConv);
        if (c) setActive({ id: c.id, other: c.other });
        else setActive({ id: initialConv, other: { id: '', username: '', displayName: '…', avatarUrl: '' } });
      }
    });
  }, [loadConversations, initialConv]);

  const openConversation = async (a: Active) => {
    setActive(a);
    const { items } = await api.messages(a.id);
    setMessages([...(items ?? [])].reverse());
    await api.markConversationRead(a.id).catch(() => {});
    setConversations((prev) => prev.map((c) => (c.id === a.id ? { ...c, unread: 0 } : c)));
    setTimeout(() => endRef.current?.scrollIntoView(), 50);
  };

  const onEvent = useCallback(
    (ev: RealtimeEvent) => {
      if (ev.type !== 'message') return;
      const msg = ev.data as Message;
      if (active && msg.conversationId === active.id) {
        setMessages((prev) => (prev.some((m) => m.id === msg.id) ? prev : [...prev, msg]));
        setTimeout(() => endRef.current?.scrollIntoView({ behavior: 'smooth' }), 30);
      }
      void loadConversations();
    },
    [active, loadConversations],
  );
  useRealtime(Boolean(user), onEvent);

  const send = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!active || !body.trim()) return;
    const text = body.trim();
    setBody('');
    const { message } = await api.sendMessage(active.id, text);
    setMessages((prev) => (prev.some((m) => m.id === message.id) ? prev : [...prev, message]));
    setTimeout(() => endRef.current?.scrollIntoView({ behavior: 'smooth' }), 30);
    void loadConversations();
  };

  if (active) {
    return (
      <div className="fixed inset-x-0 bottom-14 top-0 z-30 flex flex-col bg-white sm:static sm:z-auto sm:h-[min(42rem,calc(100dvh-6rem))]">
        <div className="flex shrink-0 items-center gap-3 border-b border-line px-4 py-3">
          <button onClick={() => setActive(null)} className="rounded-full p-1.5 hover:bg-gray-100">
            ←
          </button>
          <Avatar url={active.other.avatarUrl} name={active.other.displayName || active.other.username} size={36} />
          <div>
            <p className="font-bold leading-tight">{active.other.displayName || active.other.username}</p>
            {active.other.username && <p className="text-sm text-muted">@{active.other.username}</p>}
          </div>
        </div>

        {messages.length === 0 ? (
          <div className="flex flex-1 items-center justify-center px-6 text-center text-muted">
            <p>{t('startChat')}</p>
          </div>
        ) : (
          <div className="min-h-0 flex-1 overflow-y-auto px-4 py-3">
            {messages.map((m) => {
              const mine = m.senderId === user?.id;
              return (
                <div key={m.id} className={`mb-2 flex ${mine ? 'justify-end' : 'justify-start'}`}>
                  <div className={`max-w-[75%] rounded-2xl px-4 py-2 ${mine ? 'bg-brand text-white' : 'bg-gray-100 text-ink'}`}>
                    <p className="whitespace-pre-wrap break-words">{m.body}</p>
                    <p className={`mt-0.5 text-right text-[11px] ${mine ? 'text-white/70' : 'text-muted'}`}>{formatMessageTime(m.createdAt, locale)}</p>
                  </div>
                </div>
              );
            })}
            <div ref={endRef} />
          </div>
        )}

        <form onSubmit={send} className="flex shrink-0 items-center gap-2 border-t border-line bg-white p-3">
          <input
            value={body}
            onChange={(e) => setBody(e.target.value)}
            placeholder={t('placeholder')}
            className="input rounded-full"
            autoFocus
          />
          <button type="submit" disabled={!body.trim()} className="btn-primary shrink-0 py-2">
            {t('send')}
          </button>
        </form>
      </div>
    );
  }

  if (conversations.length === 0) {
    return (
      <>
        <PageHeader title={t('title')} />
        <div className="px-6 py-12 text-center text-muted">{t('empty')}</div>
      </>
    );
  }

  return (
    <>
    <PageHeader title={t('title')} />
    <ul>
      {conversations.map((c) => (
        <li key={c.id}>
          <button onClick={() => openConversation({ id: c.id, other: c.other })} className="flex w-full items-center gap-3 border-b border-line px-4 py-3 text-left hover:bg-gray-50">
            <Avatar url={c.other.avatarUrl} name={c.other.displayName || c.other.username} size={48} />
            <div className="min-w-0 flex-1">
              <div className="flex items-center justify-between">
                <span className="truncate font-bold">{c.other.displayName || c.other.username}</span>
                <span className="shrink-0 text-sm text-muted">{formatConversationTime(c.lastMessageAt, locale)}</span>
              </div>
              <p className="truncate text-sm text-muted">{c.lastMessage?.body ?? ''}</p>
            </div>
            {c.unread > 0 && <span className="h-2.5 w-2.5 rounded-full bg-brand" />}
          </button>
        </li>
      ))}
    </ul>
    </>
  );
}

export default function MessagesPage() {
  return (
    <AppShell rightRail={false}>
      <RequireAuth>
        <MessagesInner />
      </RequireAuth>
    </AppShell>
  );
}
