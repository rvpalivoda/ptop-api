import { Turnstile } from '@marsidev/react-turnstile';
import { cn } from '@/lib/utils';

const siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY ?? '';

export default function Captcha({ onChange }: { onChange: (token: string) => void }) {
  if (!siteKey) return null;

  return (
    <Turnstile
      siteKey={siteKey}
      options={{ theme: 'dark' }}
      className={cn('w-full')}
      onSuccess={(val) => onChange(val)}
    />
  );
}
