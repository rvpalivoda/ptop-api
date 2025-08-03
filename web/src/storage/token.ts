export interface Tokens {
  access: string;
  refresh: string;
}

const STORAGE_KEY = 'peerex_tokens';

export function saveTokens(tokens: Tokens) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
}

export function loadTokens(): Tokens | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as Tokens;
  } catch {
    return null;
  }
}

export function clearTokens() {
  localStorage.removeItem(STORAGE_KEY);
}
