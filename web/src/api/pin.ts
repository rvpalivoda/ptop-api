import { apiRequest } from './client';

export async function setPinCode(password: string, pinCode: string) {
  return apiRequest<{ status: string }>('/pin-code/set', {
    method: 'POST',
    body: JSON.stringify({ password, pin_code: pinCode }),
  });
}

