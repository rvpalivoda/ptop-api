import { apiRequest } from './client';

interface EnableResponse {
  secret: string;
  otpauth_url: string;
}

interface VerifyResponse {
  verified: boolean;
}

export async function enable2fa() {
  return apiRequest<EnableResponse>('/2fa/enable', {
    method: 'POST',
  });
}

export async function verify2fa(code: string) {
  return apiRequest<VerifyResponse>('/2fa/verify', {
    method: 'POST',
    body: JSON.stringify({ code }),
  });
}

export async function disable2fa(password: string) {
  return apiRequest<{ status: string }>('/2fa/disable', {
    method: 'POST',
    body: JSON.stringify({ password }),
  });
}
