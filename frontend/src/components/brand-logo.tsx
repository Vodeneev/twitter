type IconProps = { className?: string };

/** Bullfinch mark — no frame, red / black / white only. */
export function YapperIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 40 40" className={className} aria-hidden>
      <ellipse cx="24.5" cy="24" rx="10.5" ry="8.5" fill="#FC3F1D" />
      <circle cx="16" cy="16.5" r="7.5" fill="#21201F" />
      <ellipse cx="26.5" cy="23" rx="4.5" ry="6.5" fill="#21201F" transform="rotate(-18 26.5 23)" />
      <circle cx="17.8" cy="15" r="1.35" fill="#FFFFFF" />
      <path d="M9.2 16.8 6.5 17.8 9.2 18.8Z" fill="#FC3F1D" />
      <path d="M33.5 23.5 38.5 25 33.5 26.5Z" fill="#21201F" />
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
