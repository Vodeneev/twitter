type IconProps = { className?: string };

/** App mark: Yango yellow tile + Yandex-red yap bubble. */
export function YapperIcon({ className }: IconProps) {
  return (
    <svg viewBox="0 0 40 40" className={className} aria-hidden>
      <rect width="40" height="40" rx="11" fill="#FFCC00" />
      <path
        d="M11 12h18a3.5 3.5 0 0 1 3.5 3.5v7.5A3.5 3.5 0 0 1 29 26.5h-7.8L17 30v-3.5h-6A3.5 3.5 0 0 1 7.5 23V15.5A3.5 3.5 0 0 1 11 12Z"
        fill="#FC3F1D"
      />
      <circle cx="16.5" cy="19.5" r="1.6" fill="#fff" />
      <circle cx="20" cy="19.5" r="1.6" fill="#fff" />
      <circle cx="23.5" cy="19.5" r="1.6" fill="#fff" />
    </svg>
  );
}

export function YapperWordmark({ className = '' }: { className?: string }) {
  return (
    <span className={`font-extrabold tracking-tight ${className}`}>
      <span className="text-ya-red">Y</span>
      <span className="text-ya-yellow">A</span>
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