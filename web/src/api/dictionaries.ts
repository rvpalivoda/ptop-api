import { apiRequest } from './client';

export async function getAssets() {
  return apiRequest<any[]>('/assets');
}

export async function getCountries() {
  return apiRequest<any[]>('/countries');
}

export async function getDurations() {
  return apiRequest<any[]>('/durations');
}

export async function getPaymentMethods(country: string) {
  return apiRequest<any[]>(`/payment-methods?country=${encodeURIComponent(country)}`);
}
