'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { api } from '@/lib/api';
import { useSession } from './session-provider';
import { useRouter } from '@/i18n/navigation';

interface Props {
  username: string;
  initialFollowing: boolean;
  onChange?: (following: boolean) => void;
  small?: boolean;
}

export function FollowButton({ username, initialFollowing, onChange, small }: Props) {
  const t = useTranslations('profile');
  const { user } = useSession();
  const router = useRouter();
  const [following, setFollowing] = useState(initialFollowing);
  const [hover, setHover] = useState(false);
  const [busy, setBusy] = useState(false);

  if (user?.username === username) return null;

  const toggle = async () => {
    if (!user) {
      router.push('/login');
      return;
    }
    setBusy(true);
    try {
      if (following) {
        await api.unfollow(username);
        setFollowing(false);
        onChange?.(false);
      } else {
        await api.follow(username);
        setFollowing(true);
        onChange?.(true);
      }
    } finally {
      setBusy(false);
    }
  };

  const sizeCls = small ? 'px-4 py-1.5 text-sm' : 'px-5 py-2';

  if (following) {
    return (
      <button
        type="button"
        disabled={busy}
        onClick={toggle}
        onMouseEnter={() => setHover(true)}
        onMouseLeave={() => setHover(false)}
        className={`rounded-full border font-bold transition ${sizeCls} ${
          hover ? 'border-red-300 bg-red-50 text-red-600' : 'border-gray-300 text-ink'
        }`}
      >
        {hover ? t('unfollow') : t('following')}
      </button>
    );
  }
  return (
    <button type="button" disabled={busy} onClick={toggle} className={`rounded-full bg-ink font-bold text-white transition hover:opacity-90 ${sizeCls}`}>
      {t('follow')}
    </button>
  );
}
