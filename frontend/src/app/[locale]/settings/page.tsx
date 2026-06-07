'use client';

import { useEffect, useRef, useState } from 'react';
import { useTranslations } from 'next-intl';
import { useRouter } from '@/i18n/navigation';
import { AppShell } from '@/components/app-shell';
import { PageHeader } from '@/components/page-header';
import { RequireAuth } from '@/components/require-auth';
import { Avatar } from '@/components/avatar';
import { useSession } from '@/components/session-provider';
import { api, uploadToStorage } from '@/lib/api';

function SettingsInner() {
  const t = useTranslations('settings');
  const tn = useTranslations('nav');
  const router = useRouter();
  const { user, setUser, refresh, logout } = useSession();
  const [displayName, setDisplayName] = useState('');
  const [bio, setBio] = useState('');
  const [location, setLocation] = useState('');
  const [website, setWebsite] = useState('');
  const [avatarKey, setAvatarKey] = useState<string | undefined>();
  const [avatarPreview, setAvatarPreview] = useState<string | undefined>();
  const [headerKey, setHeaderKey] = useState<string | undefined>();
  const [headerPreview, setHeaderPreview] = useState<string | undefined>();
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);
  const avatarRef = useRef<HTMLInputElement>(null);
  const headerRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (user) {
      setDisplayName(user.displayName);
      setBio(user.bio);
      setLocation(user.location);
      setWebsite(user.website);
      setAvatarPreview(user.avatarUrl || undefined);
      setHeaderPreview(user.headerUrl || undefined);
    }
  }, [user]);

  const upload = async (file: File, kind: 'avatar' | 'header') => {
    const { uploadUrl, key, publicUrl } = await api.presign(file.type, kind);
    await uploadToStorage(uploadUrl, file);
    if (kind === 'avatar') {
      setAvatarKey(key);
      setAvatarPreview(publicUrl);
    } else {
      setHeaderKey(key);
      setHeaderPreview(publicUrl);
    }
  };

  const save = async () => {
    setBusy(true);
    setSaved(false);
    try {
      const { user: updated } = await api.updateMe({ displayName, bio, location, website, avatarKey, headerKey });
      setUser(updated);
      await refresh();
      setSaved(true);
    } finally {
      setBusy(false);
    }
  };

  if (!user) return null;

  return (
    <div className="pb-10">
      <div className="relative z-0 h-44 w-full overflow-hidden bg-line">
        {headerPreview && (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={headerPreview} alt="" className="h-full w-full object-cover object-top" />
        )}
        <button onClick={() => headerRef.current?.click()} className="absolute inset-0 z-10 flex items-center justify-center bg-black/30 text-sm font-semibold text-white">
          {t('header')}
        </button>
        <input ref={headerRef} type="file" accept="image/*" hidden onChange={(e) => e.target.files?.[0] && upload(e.target.files[0], 'header')} />
      </div>
      <div className="relative z-20 px-4">
        <button onClick={() => avatarRef.current?.click()} className="-mt-12 block rounded-full">
          <Avatar url={avatarPreview} name={displayName || user.username} size={96} className="border-4 border-white bg-white" />
        </button>
        <input ref={avatarRef} type="file" accept="image/*" hidden onChange={(e) => e.target.files?.[0] && upload(e.target.files[0], 'avatar')} />

        <div className="mt-4 flex flex-col gap-4">
          <label className="text-sm font-semibold text-muted">
            {t('displayName')}
            <input className="input mt-1" value={displayName} maxLength={50} onChange={(e) => setDisplayName(e.target.value)} />
          </label>
          <label className="text-sm font-semibold text-muted">
            {t('bio')}
            <textarea className="input mt-1" rows={3} value={bio} maxLength={160} onChange={(e) => setBio(e.target.value)} />
          </label>
          <label className="text-sm font-semibold text-muted">
            {t('location')}
            <input className="input mt-1" value={location} onChange={(e) => setLocation(e.target.value)} />
          </label>
          <label className="text-sm font-semibold text-muted">
            {t('website')}
            <input className="input mt-1" value={website} onChange={(e) => setWebsite(e.target.value)} />
          </label>
          <div className="flex items-center gap-3">
            <button onClick={save} disabled={busy} className="btn-primary">
              {t('save')}
            </button>
            {saved && <span className="text-green-600">{t('saved')}</span>}
          </div>
          <button
            type="button"
            onClick={async () => {
              await logout();
              router.push('/');
            }}
            className="btn-outline mt-2 w-full border-red-200 text-red-600 hover:bg-red-50"
          >
            {tn('logout')}
          </button>
        </div>
      </div>
    </div>
  );
}

export default function SettingsPage() {
  const t = useTranslations('settings');
  return (
    <AppShell rightRail={false}>
      <PageHeader title={t('title')} back />
      <RequireAuth>
        <SettingsInner />
      </RequireAuth>
    </AppShell>
  );
}
