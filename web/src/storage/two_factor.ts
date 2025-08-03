export const TWO_FACTOR_KEY = 'peerex_two_factor_enabled';

export function saveTwoFactorEnabled(value: boolean) {
  localStorage.setItem(TWO_FACTOR_KEY, JSON.stringify(value));
}

export function loadTwoFactorEnabled(): boolean {
  const raw = localStorage.getItem(TWO_FACTOR_KEY);
  if (!raw) return false;
  try {
    return JSON.parse(raw) as boolean;
  } catch {
    return false;
  }
}

export function clearTwoFactorEnabled() {
  localStorage.removeItem(TWO_FACTOR_KEY);
}
