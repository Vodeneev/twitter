'use client';

import { useCallback, useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useLocale, useTranslations } from 'next-intl';
import { AppShell } from '@/components/app-shell';
import { PageHeader, Tabs } from '@/components/page-header';
import { Timeline } from '@/components/timeline';
import { Avatar } from '@/components/avatar';
import { FollowButton } from '@/components/follow-button';
import { useSession } from '@/components/session-provider';
import { Link, useRouter } from '@/i18n/navigation';
import { api, ApiError } from '@/lib/api';
import { formatDate } from '@/lib/format';
import type { User } from '@/lib/types';

type Tab = 'yaps' | 'replies' | 'media' | 'likes';

export default function ProfilePage() {
  const t = useTranslations('profile');
  const locale = useLocale();
  const router = useRouter();
  const params = useParams();
  const username = decodeURIComponent(String(params.username));
  const { user: me } = useSession();
  const [profile, setProfile] = useState<User | null>(null);
  const [notFound, setNotFound] = useState(false);
  const [tab, setTab] = useState<Tab>('yaps');

  useEffect(() => {
    setProfile(null);
    setNotFound(false);
    api
      .getProfile(username)
      .then((r) => setProfile(r.user))
      .catch((e) => {
        if (e instanceof ApiError && e.status === 404) setNotFound(true);
      });
  }, [username]);

  const fetcher = useCallback(
    (cursor?: string) => {
      switch (tab) {
        case 'replies':
          return api.userReplies(username, cursor);
        case 'media':
          return api.userMedia(username, cursor);
        case 'likes':
          return api.userLikes(username, cursor);
        default:
          return api.userYaps(username, cursor);
      }
    },
    [tab, username],
  );

  if (notFound) {
    return (
      <AppShell>
        <PageHeader title={t('notFound')} back />
        <p className="px-6 py-12 text-center text-muted">@{username}</p>
      </AppShell>
    );
  }

  const isMe = me?.username.toLowerCase() === username.toLowerCase();
  const name = profile?.displayName || username;

  const openDM = async () => {
    const { conversationId } = await api.openConversation(username);
    router.push(`/messages?c=${conversationId}`);
  };

  return (
    <AppShell>
      <PageHeader title={name} subtitle={profile ? `${profile.yapsCount} ${t('yaps')}` : undefined} back />

      {profile && (
        <>
          <div className="relative z-0 h-44 w-full overflow-hidden bg-line">
            {profile.headerUrl && (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={profile.headerUrl} alt="" className="h-full w-full object-cover object-top" />
            )}
          </div>
          <div className="relative z-20 px-4 pb-3">
            <div className="flex items-end justify-between">
              <div className="-mt-12">
                <Avatar url={profile.avatarUrl} name={name} size={96} className="border-4 border-white bg-white" />
              </div>
              <div className="mt-3 flex gap-2">
                {isMe ? (
                  <Link href="/settings" className="btn-outline py-1.5">
                    {t('editProfile')}
                  </Link>
                ) : (
                  <>
                    {me && (
                      <button onClick={openDM} className="btn-outline py-1.5">
                        {t('message')}
                      </button>
                    )}
                    <FollowButton username={profile.username} initialFollowing={profile.following} />
                  </>
                )}
              </div>
            </div>

            <div className="mt-2">
              <h2 className="text-xl font-extrabold">{name}</h2>
              <p className="text-muted">@{profile.username}</p>
            </div>
            {profile.bio && <p className="mt-2 whitespace-pre-wrap">{profile.bio}</p>}
            <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted">
              {profile.location && <span>📍 {profile.location}</span>}
              {profile.website && (
                <a href={profile.website} target="_blank" rel="noreferrer" className="text-brand hover:underline">
                  🔗 {profile.website.replace(/^https?:\/\//, '')}
                </a>
              )}
              <span>📅 {t('joined', { date: formatDate(profile.createdAt, locale) })}</span>
            </div>
            <div className="mt-2 flex gap-4 text-sm">
              <Link href={`/${profile.username}/following`} className="hover:underline">
                <span className="font-bold text-ink">{profile.followingCount}</span> <span className="text-muted">{t('followingCount')}</span>
              </Link>
              <Link href={`/${profile.username}/followers`} className="hover:underline">
                <span className="font-bold text-ink">{profile.followersCount}</span> <span className="text-muted">{t('followers')}</span>
              </Link>
            </div>
          </div>

          <Tabs<Tab>
            tabs={[
              { id: 'yaps', label: t('tabs.yaps') },
              { id: 'replies', label: t('tabs.replies') },
              { id: 'media', label: t('tabs.media') },
              { id: 'likes', label: t('tabs.likes') },
            ]}
            active={tab}
            onChange={setTab}
          />
          <Timeline fetcher={fetcher} emptyText="—" reloadToken={['yaps', 'replies', 'media', 'likes'].indexOf(tab)} />
        </>
      )}
    </AppShell>
  );
}
