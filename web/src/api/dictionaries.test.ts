import { describe, it, expect, vi } from 'vitest';
import { getPaymentMethods } from './dictionaries';
import { apiRequest } from './client';

vi.mock('./client', () => ({
  apiRequest: vi.fn(),
}));

describe('dictionaries api', () => {
  it('запрашивает методы оплаты без страны', async () => {
    const apiReqMock = apiRequest as unknown as ReturnType<typeof vi.fn>;
    apiReqMock.mockResolvedValue([]);
    await getPaymentMethods();
    expect(apiRequest).toHaveBeenCalledWith('/payment-methods');
  });

  it('запрашивает методы оплаты с указанной страной', async () => {
    const apiReqMock = apiRequest as unknown as ReturnType<typeof vi.fn>;
    apiReqMock.mockResolvedValue([]);
    await getPaymentMethods('RU');
    expect(apiRequest).toHaveBeenCalledWith('/payment-methods?country=RU');
  });
});
