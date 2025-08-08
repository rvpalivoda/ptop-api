
interface FilterPanelProps {
  filters: {
    fromAsset: string;
    toAsset: string;
    minAmount: string;
    maxAmount: string;
    paymentMethod: string;
  };
  onFiltersChange: (filters: {
    fromAsset: string;
    toAsset: string;
    minAmount: string;
    maxAmount: string;
    paymentMethod: string;
  }) => void;
  activeTab: 'buy' | 'sell';
  onTabChange: (tab: 'buy' | 'sell') => void;
  onCreate: () => void;
}

const assets = [
  { value: 'all', label: 'Все' },
  { value: 'BTC', label: 'BTC' },
  { value: 'ETH', label: 'ETH' },
  { value: 'USDT', label: 'USDT' },
  { value: 'BNB', label: 'BNB' }
];

const paymentMethods = [
  { value: 'all', label: 'Все способы' },
  { value: 'sberbank', label: 'Сбербанк' },
  { value: 'tinkoff', label: 'Тинькофф' },
  { value: 'alfa', label: 'Альфа-Банк' },
  { value: 'qiwi', label: 'QIWI' },
  { value: 'yandex', label: 'ЮMoney' }
];

export const FilterPanel = ({
  filters,
  onFiltersChange,
  activeTab,
  onTabChange,
  onCreate
}: FilterPanelProps) => {
  return (
    <div className="bg-gray-800 rounded-lg p-4 border border-gray-700 mb-3">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        <select
          data-testid="from-asset"
          value={filters.fromAsset}
          onChange={(e) => onFiltersChange({ ...filters, fromAsset: e.target.value })}
          className="bg-gray-700 border border-gray-600 rounded-md px-2 py-1 text-white focus:outline-none"
        >
          {assets.map((a) => (
            <option key={a.value} value={a.value}>
              {a.label}
            </option>
          ))}
        </select>

        <select
          data-testid="to-asset"
          value={filters.toAsset}
          onChange={(e) => onFiltersChange({ ...filters, toAsset: e.target.value })}
          className="bg-gray-700 border border-gray-600 rounded-md px-2 py-1 text-white focus:outline-none"
        >
          {assets.map((a) => (
            <option key={a.value} value={a.value}>
              {a.label}
            </option>
          ))}
        </select>

        <input
          data-testid="min-amount"
          type="number"
          placeholder="Мин"
          value={filters.minAmount}
          onChange={(e) => onFiltersChange({ ...filters, minAmount: e.target.value })}
          className="w-24 bg-gray-700 border border-gray-600 rounded-md px-2 py-1 text-white placeholder-gray-400 focus:outline-none"
        />

        <input
          data-testid="max-amount"
          type="number"
          placeholder="Макс"
          value={filters.maxAmount}
          onChange={(e) => onFiltersChange({ ...filters, maxAmount: e.target.value })}
          className="w-24 bg-gray-700 border border-gray-600 rounded-md px-2 py-1 text-white placeholder-gray-400 focus:outline-none"
        />

        <select
          data-testid="payment-method"
          value={filters.paymentMethod}
          onChange={(e) => onFiltersChange({ ...filters, paymentMethod: e.target.value })}
          className="bg-gray-700 border border-gray-600 rounded-md px-2 py-1 text-white focus:outline-none"
        >
          {paymentMethods.map((m) => (
            <option key={m.value} value={m.value}>
              {m.label}
            </option>
          ))}
        </select>

        <div className="flex rounded-md overflow-hidden border border-gray-600">
          <button
            data-testid="buy-tab"
            onClick={() => onTabChange('buy')}
            className={`px-3 py-1 text-sm ${
              activeTab === 'buy' ? 'bg-green-600 text-white' : 'text-gray-300'
            }`}
          >
            Купить
          </button>
          <button
            data-testid="sell-tab"
            onClick={() => onTabChange('sell')}
            className={`px-3 py-1 text-sm ${
              activeTab === 'sell' ? 'bg-red-600 text-white' : 'text-gray-300'
            }`}
          >
            Продать
          </button>
        </div>
      </div>

      <div className="mt-4">
        <button
          onClick={onCreate}
          className="bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white px-6 py-3 rounded-lg font-medium transition-all duration-200 transform hover:scale-105"
        >
          + Создать объявление
        </button>
      </div>
    </div>
  );
};

