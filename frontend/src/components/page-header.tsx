'use client';

import { useRouter } from '@/i18n/navigation';

export function PageHeader({ title, subtitle, back }: { title: string; subtitle?: string; back?: boolean }) {
  const router = useRouter();
  return (
    <div className="sticky top-0 z-10 flex items-center gap-4 border-b border-line bg-white/80 px-4 py-3 backdrop-blur">
      {back && (
        <button onClick={() => router.back()} className="rounded-full p-1.5 hover:bg-gray-100" aria-label="back">
          <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth={2}>
            <path d="M15 5l-7 7 7 7" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </button>
      )}
      <div>
        <h1 className="text-xl font-extrabold leading-tight">{title}</h1>
        {subtitle && <p className="text-sm text-muted">{subtitle}</p>}
      </div>
    </div>
  );
}

export function Tabs<T extends string>({ tabs, active, onChange }: { tabs: { id: T; label: string }[]; active: T; onChange: (id: T) => void }) {
  return (
    <div className="sticky top-0 z-10 flex border-b border-line bg-white/80 backdrop-blur">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => onChange(tab.id)}
          className="relative flex-1 py-4 text-center font-semibold transition hover:bg-gray-50"
        >
          <span className={active === tab.id ? 'text-ink' : 'text-muted'}>{tab.label}</span>
          {active === tab.id && <span className="absolute bottom-0 left-1/2 h-1 w-14 -translate-x-1/2 rounded-full bg-brand" />}
        </button>
      ))}
    </div>
  );
}
