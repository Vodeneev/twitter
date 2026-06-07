'use client';

import { useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { Link, useRouter } from '@/i18n/navigation';
import { api } from '@/lib/api';
import type { Yap } from '@/lib/types';
import { timeAgo } from '@/lib/format';
import { useSession } from './session-provider';
import { Avatar } from './avatar';
import { YapContent } from './yap-content';
import { MediaGrid } from './media-grid';
import { YapComposer } from './yap-composer';
import { BookmarkIcon, HeartIcon, QuoteIcon, ReplyIcon, RepostIcon, TrashIcon } from './icons';

interface Props {
  yap: Yap;
  onChange?: (yap: Yap) => void;
  onDelete?: (id: string) => void;
  showThreadLine?: boolean;
}

export function YapCard({ yap, onChange, onDelete, showThreadLine }: Props) {
  const t = useTranslations('yap');
  const locale = useLocale();
  const router = useRouter();
  const { user } = useSession();
  const [state, setState] = useState<Yap>(yap);
  const [repostMenu, setRepostMenu] = useState(false);
  const [quoteOpen, setQuoteOpen] = useState(false);

  const update = (patch: Partial<Yap>) => {
    const next = { ...state, ...patch };
    setState(next);
    onChange?.(next);
  };

  const requireAuth = () => {
    if (!user) {
      router.push('/login');
      return false;
    }
    return true;
  };

  const toggleLike = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!requireAuth()) return;
    const liked = !state.liked;
    update({ liked, likesCount: state.likesCount + (liked ? 1 : -1) });
    try {
      liked ? await api.like(state.id) : await api.unlike(state.id);
    } catch {
      update({ liked: !liked, likesCount: state.likesCount });
    }
  };

  const toggleRepost = async () => {
    if (!requireAuth()) return;
    setRepostMenu(false);
    const reposted = !state.reposted;
    update({ reposted, repostsCount: state.repostsCount + (reposted ? 1 : -1) });
    try {
      reposted ? await api.repost(state.id) : await api.unrepost(state.id);
    } catch {
      update({ reposted: !reposted, repostsCount: state.repostsCount });
    }
  };

  const toggleBookmark = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!requireAuth()) return;
    const bookmarked = !state.bookmarked;
    update({ bookmarked, bookmarksCount: state.bookmarksCount + (bookmarked ? 1 : -1) });
    try {
      bookmarked ? await api.bookmark(state.id) : await api.unbookmark(state.id);
    } catch {
      update({ bookmarked: !bookmarked, bookmarksCount: state.bookmarksCount });
    }
  };

  const remove = async (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm(t('deleteConfirm'))) return;
    try {
      await api.deleteYap(state.id);
      onDelete?.(state.id);
    } catch {
      /* ignore */
    }
  };

  const goDetail = () => router.push(`/yap/${state.id}`);
  const name = state.author.displayName || state.author.username;
  const isOwner = user?.id === state.author.id;

  return (
    <article onClick={goDetail} className="relative cursor-pointer border-b border-line px-4 py-3 transition hover:bg-gray-50">
      {state.repostedBy && (
        <div className="mb-1 flex items-center gap-2 pl-6 text-sm font-semibold text-muted">
          <RepostIcon className="h-4 w-4" />
          <span>
            {state.repostedBy.displayName || state.repostedBy.username} {t('reposted')}
          </span>
        </div>
      )}
      <div className="flex gap-3">
        <div className="flex flex-col items-center">
          <Link href={`/${state.author.username}`} onClick={(e) => e.stopPropagation()}>
            <Avatar url={state.author.avatarUrl} name={name} size={44} />
          </Link>
          {showThreadLine && <div className="mt-1 w-0.5 flex-1 bg-line" />}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1 text-[15px]">
            <Link href={`/${state.author.username}`} onClick={(e) => e.stopPropagation()} className="truncate font-bold hover:underline">
              {name}
            </Link>
            <span className="truncate text-muted">@{state.author.username}</span>
            <span className="text-muted">·</span>
            <span className="shrink-0 text-muted">{timeAgo(state.createdAt, locale)}</span>
            {isOwner && (
              <button onClick={remove} className="ml-auto rounded-full p-1.5 text-muted hover:bg-red-50 hover:text-red-500" aria-label={t('delete')}>
                <TrashIcon className="h-4 w-4" />
              </button>
            )}
          </div>

          {state.content && (
            <div className="mt-0.5 text-[15px] leading-snug">
              <YapContent content={state.content} />
            </div>
          )}

          <MediaGrid media={state.media} />

          {state.quoteOf && (
            <Link
              href={`/yap/${state.quoteOf.id}`}
              onClick={(e) => e.stopPropagation()}
              className="mt-3 block rounded-2xl border border-line p-3 hover:bg-gray-100"
            >
              <div className="flex items-center gap-1 text-sm">
                <Avatar url={state.quoteOf.author.avatarUrl} name={state.quoteOf.author.displayName || state.quoteOf.author.username} size={20} />
                <span className="font-bold">{state.quoteOf.author.displayName || state.quoteOf.author.username}</span>
                <span className="text-muted">@{state.quoteOf.author.username}</span>
              </div>
              <div className="mt-1 text-[15px] leading-snug">
                <YapContent content={state.quoteOf.content} />
              </div>
              <MediaGrid media={state.quoteOf.media} />
            </Link>
          )}

          {/* Action bar */}
          <div className="mt-2 flex max-w-md items-center justify-between text-muted">
            <button onClick={(e) => { e.stopPropagation(); goDetail(); }} className="group flex items-center gap-1.5 hover:text-brand">
              <span className="rounded-full p-1.5 group-hover:bg-brand/10">
                <ReplyIcon className="h-[18px] w-[18px]" />
              </span>
              <span className="text-sm">{state.repliesCount || ''}</span>
            </button>

            <div className="relative">
              <button onClick={(e) => { e.stopPropagation(); setRepostMenu((v) => !v); }} className={`group flex items-center gap-1.5 hover:text-green-600 ${state.reposted ? 'text-green-600' : ''}`}>
                <span className="rounded-full p-1.5 group-hover:bg-green-600/10">
                  <RepostIcon className="h-[18px] w-[18px]" />
                </span>
                <span className="text-sm">{state.repostsCount || ''}</span>
              </button>
              {repostMenu && (
                <div onClick={(e) => e.stopPropagation()} className="absolute z-10 mt-1 w-40 overflow-hidden rounded-xl border border-line bg-white shadow-lg">
                  <button onClick={toggleRepost} className="flex w-full items-center gap-2 px-4 py-2.5 text-left hover:bg-gray-100">
                    <RepostIcon className="h-4 w-4" /> {state.reposted ? t('repost') + ' ✓' : t('repost')}
                  </button>
                  <button onClick={() => { setRepostMenu(false); setQuoteOpen(true); }} className="flex w-full items-center gap-2 px-4 py-2.5 text-left hover:bg-gray-100">
                    <QuoteIcon className="h-4 w-4" /> {t('quote')}
                  </button>
                </div>
              )}
            </div>

            <button onClick={toggleLike} className={`group flex items-center gap-1.5 hover:text-pink-600 ${state.liked ? 'text-pink-600' : ''}`}>
              <span className="rounded-full p-1.5 group-hover:bg-pink-600/10">
                <HeartIcon className="h-[18px] w-[18px]" filled={state.liked} />
              </span>
              <span className="text-sm">{state.likesCount || ''}</span>
            </button>

            <button onClick={toggleBookmark} className={`group flex items-center gap-1.5 hover:text-brand ${state.bookmarked ? 'text-brand' : ''}`}>
              <span className="rounded-full p-1.5 group-hover:bg-brand/10">
                <BookmarkIcon className="h-[18px] w-[18px]" filled={state.bookmarked} />
              </span>
            </button>
          </div>
        </div>
      </div>

      {quoteOpen && (
        <div onClick={(e) => { e.stopPropagation(); setQuoteOpen(false); }} className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 p-4 pt-20">
          <div onClick={(e) => e.stopPropagation()} className="w-full max-w-feed rounded-2xl bg-white p-2 shadow-xl">
            <YapComposer
              quoteOfId={state.id}
              autoFocus
              onCreated={() => {
                setQuoteOpen(false);
                update({ repostsCount: state.repostsCount });
              }}
            />
            <div className="mx-4 mb-2 rounded-2xl border border-line p-3 text-sm">
              <span className="font-bold">{name}</span> <span className="text-muted">@{state.author.username}</span>
              <div className="mt-1">{state.content}</div>
            </div>
          </div>
        </div>
      )}
    </article>
  );
}
