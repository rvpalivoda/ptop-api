import { apiRequest } from './client';

export async function setPinCode(password: string, pinCode: string) {
  return apiRequest<{ status: string }>('/auth/pincode', {
    method: 'POST',
    body: JSON.stringify({ password, pincode: pinCode }),
  });
}

