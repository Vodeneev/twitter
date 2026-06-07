import { initialsAvatar } from '@/lib/format';

interface AvatarProps {
  url?: string;
  name: string;
  size?: number;
  className?: string;
}

export function Avatar({ url, name, size = 48, className = '' }: AvatarProps) {
  const dim = { width: size, height: size };
  if (url) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={url} alt={name} style={dim} className={`rounded-full object-cover bg-line ${className}`} />;
  }
  return (
    <div
      style={dim}
      className={`flex items-center justify-center rounded-full bg-brand font-bold text-white ${className}`}
    >
      <span style={{ fontSize: size * 0.42 }}>{initialsAvatar(name)}</span>
    </div>
  );
}
