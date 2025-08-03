import { apiRequest } from './client';

interface EnableResponse {
  secret: string;
  url: string;
}

export async function enable2fa(password: string) {
  return apiRequest<EnableResponse>('/auth/2fa/enable', {
    method: 'POST',
    body: JSON.stringify({ password }),
  });
}

