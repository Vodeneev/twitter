'use client';

import { useState } from 'react';
import { useLocale, useTranslations } from 'next-intl';
import { AuthCard } from '@/components/auth-card';
import { Link, useRouter } from '@/i18n/navigation';
import { useSession } from '@/components/session-provider';
import { api, ApiError } from '@/lib/api';

export default function LoginPage() {
  const t = useTranslations('auth');
  const te = useTranslations('auth.errors');
  const locale = useLocale();
  const router = useRouter();
  const { setUser } = useSession();

  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(true);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError('');
    setNotice('');
    try {
      const { user } = await api.login({ identifier, password, rememberMe });
      setUser(user);
      router.push('/');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'email_not_verified') {
          await api.resendVerification(identifier.includes('@') ? identifier : '', locale).catch(() => {});
          setNotice(t('emailNotVerified'));
        } else {
          setError(te.has(err.code) ? te(err.code) : te('internal_error'));
        }
      } else {
        setError(te('internal_error'));
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <AuthCard title={t('loginTitle')}>
      <form onSubmit={submit} className="flex flex-col gap-4">
        {error && <p className="alert-error">{error}</p>}
        {notice && <p className="alert-info">{notice}</p>}
        <input className="input" placeholder={t('identifier')} value={identifier} onChange={(e) => setIdentifier(e.target.value)} autoComplete="username" required />
        <input className="input" type="password" placeholder={t('password')} value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" required />
        <label className="flex items-center gap-2 text-sm text-muted">
          <input type="checkbox" checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)} />
          {t('rememberMe')}
        </label>
        <button type="submit" disabled={busy} className="btn-primary">
          {t('loginButton')}
        </button>
      </form>
      <div className="mt-4 flex justify-between text-sm">
        <Link href="/forgot-password" className="text-brand hover:underline">
          {t('forgotPassword')}
        </Link>
        <span className="text-muted">
          {t('noAccount')}{' '}
          <Link href="/register" className="text-brand hover:underline">
            {t('registerButton')}
          </Link>
        </span>
      </div>
    </AuthCard>
  );
}
