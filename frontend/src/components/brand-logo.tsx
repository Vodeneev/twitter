type IconProps = { className?: string };

/** Stylized bullfinch (снегирь) — red, white, black. */
export function YapperIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 40 40" className={className} aria-hidden>
      <rect width="40" height="40" rx="11" fill="#FFFFFF" stroke="#21201F" strokeWidth="1.25" />
      <path d="M10 27h20" stroke="#21201F" strokeWidth="2" strokeLinecap="round" />
      <ellipse cx="23" cy="22.5" rx="7.5" ry="6.5" fill="#FC3F1D" />
      <circle cx="17.5" cy="17" r="5.5" fill="#21201F" />
      <circle cx="18.8" cy="18.2" r="1.3" fill="#FFFFFF" />
      <circle cx="16.2" cy="15.8" r="0.9" fill="#FFFFFF" />
      <path d="M13.5 17.2 10.5 18.5 13.8 19.2Z" fill="#FC3F1D" />
      <path d="M27 21.5c2.5 1.2 3.8 3.5 3.2 6.2" fill="#21201F" />
      <path d="M21 27.5c-1.5 1-3.2 1.2-4.8.6" stroke="#21201F" strokeWidth="1.5" strokeLinecap="round" fill="none" />
    </svg>
  );
}

export function YapperWordmark({ className = '' }: { className?: string }) {
  return (
    <span className={`font-extrabold tracking-tight ${className}`}>
      <span className="text-brand">YA</span>
      <span className="text-ink">pper</span>
    </span>
  );
}

export function YapperLogo({
  iconClassName = 'h-9 w-9',
  wordmarkClassName = 'text-2xl',
  showWordmark = true,
}: {
  iconClassName?: string;
  wordmarkClassName?: string;
  showWordmark?: boolean;
}) {
  return (
    <span className="inline-flex items-center gap-2">
      <YapperIcon className={iconClassName} />
      {showWordmark ? <YapperWordmark className={wordmarkClassName} /> : null}
    </span>
  );
}
