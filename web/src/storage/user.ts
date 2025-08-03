export interface UserInfo {
  username: string;
  twofaEnabled: boolean;
  pinCodeSet: boolean;
}

const STORAGE_KEY = 'peerex_user_info';

export function saveUserInfo(info: UserInfo) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(info));
}

export function loadUserInfo(): UserInfo | null {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as UserInfo;
  } catch {
    return null;
  }
}

export function clearUserInfo() {
  localStorage.removeItem(STORAGE_KEY);
}
