'use client';

import { useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { AuthCard } from '@/components/auth-card';
import { Link } from '@/i18n/navigation';
import { api, ApiError } from '@/lib/api';

export default function RegisterPage() {
  const t = useTranslations('auth');
  const te = useTranslations('auth.errors');
  const locale = useLocale();

  const [displayName, setDisplayName] = useState('');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [done, setDone] = useState(false);
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    try {
      await api.register({ displayName, username, email, password, locale });
      setDone(true);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'validation' && err.message) setError(err.message);
        else setError(te.has(err.code) ? te(err.code) : te('internal_error'));
      } else {
        setError(te('internal_error'));
      }
    } finally {
      setBusy(false);
    }
  };

  if (done) {
    return (
      <AuthCard title={t('registerTitle')}>
        <p className="rounded-md bg-ya-yellow/15 px-3 py-3 text-center text-brand">{t('checkEmail')}</p>
        <div className="mt-4 text-center text-sm">
          <Link href="/login" className="text-brand hover:underline">
            {t('loginButton')}
          </Link>
        </div>
      </AuthCard>
    );
  }

  return (
    <AuthCard title={t('registerTitle')}>
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && <p className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-600">{error}</p>}
        <input className="input" placeholder={t('displayName')} value={displayName} onChange={(e) => setDisplayName(e.target.value)} required />
        <input className="input" placeholder={t('username')} value={username} onChange={(e) => setUsername(e.target.value)} pattern="[a-zA-Z0-9_]{3,20}" required />
        <input className="input" type="email" placeholder={t('email')} value={email} onChange={(e) => setEmail(e.target.value)} required />
        <input className="input" type="password" placeholder={t('password')} value={password} onChange={(e) => setPassword(e.target.value)} minLength={8} required />
        <button type="submit" disabled={busy} className="btn-primary">
          {t('registerButton')}
        </button>
      </form>
      <div className="mt-4 text-center text-sm text-muted">
        {t('haveAccount')}{' '}
        <Link href="/login" className="text-brand hover:underline">
          {t('loginButton')}
        </Link>
      </div>
    </AuthCard>
  );
}
