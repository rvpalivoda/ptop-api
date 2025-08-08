import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('@/api/dictionaries', () => ({
  getAssets: vi.fn().mockResolvedValue([
    { id: 'BTC', name: 'BTC' },
    { id: 'ETH', name: 'ETH' },
  ]),
  getPaymentMethods: vi.fn().mockResolvedValue([
    { id: 'pm1', name: 'Сбербанк' },
    { id: 'pm2', name: 'Тинькофф' },
  ]),
}));

import { FilterPanel } from './FilterPanel';

describe('FilterPanel', () => {
  const baseFilters = {
    fromAsset: 'all',
    toAsset: 'all',
    minAmount: '',
    maxAmount: '',
    paymentMethod: 'all'
  };

  it('вызывает onFiltersChange при смене актива', async () => {
    const onFiltersChange = vi.fn();
    render(
      <FilterPanel
        filters={baseFilters}
        onFiltersChange={onFiltersChange}
        activeTab="buy"
        onTabChange={() => {}}
        onCreate={() => {}}
      />
    );

    const select = (await screen.findAllByTestId('from-asset'))[0];
    fireEvent.change(select, {
      target: { value: 'BTC' }
    });

    expect(onFiltersChange).toHaveBeenCalledWith({
      ...baseFilters,
      fromAsset: 'BTC'
    });
  });

  it('переключает тип сделки', async () => {
    const onTabChange = vi.fn();
    render(
      <FilterPanel
        filters={baseFilters}
        onFiltersChange={() => {}}
        activeTab="buy"
        onTabChange={onTabChange}
        onCreate={() => {}}
      />
    );

    await screen.findAllByTestId('from-asset');
    const sellBtn = screen.getAllByTestId('sell-tab').pop();
    if (sellBtn) {
      fireEvent.click(sellBtn);
    }
    expect(onTabChange).toHaveBeenCalledWith('sell');
  });
});
