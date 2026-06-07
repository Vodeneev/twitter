'use client';

import { Link } from '@/i18n/navigation';
import { YapperLogo } from './brand-logo';

export function AuthCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-md rounded-2xl border border-line bg-white p-8 shadow-sm">
        <Link href="/" className="mb-4 flex justify-center">
          <YapperLogo iconClassName="h-10 w-10" wordmarkClassName="text-2xl" />
        </Link>
        <h1 className="mb-6 text-center text-2xl font-extrabold">{title}</h1>
        {children}
      </div>
    </div>
  );
}
