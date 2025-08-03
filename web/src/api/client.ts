import { loadTokens, saveTokens, clearTokens } from '@/storage/token';
import { clearUserInfo } from '@/storage/user';
import { refresh as refreshTokens } from './auth';

const BASE_URL = (
  import.meta.env.VITE_API_BASE_URL ?? '/api/v1'
).replace(/\/$/, '');

function forceLogout() {
  clearTokens();
  clearUserInfo();
  window.location.assign('/login');
}

export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {},
  isRetry = false,
): Promise<T> {
  const tokens = loadTokens();
  const headers = new Headers(options.headers);
  headers.set('Content-Type', 'application/json');
  if (tokens?.access) {
    headers.set('Authorization', `Bearer ${tokens.access}`);
  }

  const response = await fetch(`${BASE_URL}${endpoint}`, {
    ...options,
    headers,
  });

  if (
    response.status === 401 &&
    !isRetry &&
    endpoint !== '/auth/refresh' &&
    endpoint !== '/auth/login'
  ) {
    try {
      const newTokens = await refreshTokens();
      saveTokens(newTokens);
      return apiRequest<T>(endpoint, options, true);
    } catch (err) {
      forceLogout();
      throw new Error('Unauthorized');
    }
  }

  if (endpoint === '/auth/refresh' && !response.ok) {
    forceLogout();
  }

  if (!response.ok) {
    let message = `API error: ${response.status}`;
    try {
      const data = await response.json();
      if (data && typeof (data as any).detail === 'string') {
        message = (data as any).detail;
      }
    } catch {
      // ignore json parse errors and use default message
    }
    throw new Error(message);
  }
  return (await response.json()) as T;
}
