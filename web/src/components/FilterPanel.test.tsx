import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { FilterPanel } from './FilterPanel';

describe('FilterPanel', () => {
  const baseFilters = {
    fromAsset: 'all',
    toAsset: 'all',
    minAmount: '',
    maxAmount: '',
    paymentMethod: 'all'
  };

  it('вызывает onFiltersChange при смене актива', () => {
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

    fireEvent.change(screen.getByTestId('from-asset'), {
      target: { value: 'BTC' }
    });

    expect(onFiltersChange).toHaveBeenCalledWith({
      ...baseFilters,
      fromAsset: 'BTC'
    });
  });

  it('переключает тип сделки', () => {
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

    const sellBtn = screen.getAllByTestId('sell-tab').pop();
    if (sellBtn) {
      fireEvent.click(sellBtn);
    }
    expect(onTabChange).toHaveBeenCalledWith('sell');
  });
});
