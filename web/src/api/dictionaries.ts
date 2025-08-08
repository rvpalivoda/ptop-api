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

export async function getPaymentMethods(country?: string) {
  const endpoint = country
    ? `/payment-methods?country=${encodeURIComponent(country)}`
    : '/payment-methods';
  return apiRequest<any[]>(endpoint);
}
