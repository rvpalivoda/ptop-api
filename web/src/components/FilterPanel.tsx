
import { Search, Filter } from 'lucide-react';

interface FilterPanelProps {
  filters: {
    currency: string;
    paymentMethod: string;
    amount: string;
  };
  onFiltersChange: (filters: any) => void;
  activeTab: 'buy' | 'sell';
  onTabChange: (tab: 'buy' | 'sell') => void;
  onCreate: () => void;
}

export const FilterPanel = ({
  filters,
  onFiltersChange,
  activeTab,
  onTabChange,
  onCreate
}: FilterPanelProps) => {
  const currencies = [
    { value: 'all', label: 'Все валюты' },
    { value: 'BTC', label: 'Bitcoin (BTC)' },
    { value: 'ETH', label: 'Ethereum (ETH)' },
    { value: 'USDT', label: 'Tether (USDT)' },
    { value: 'BNB', label: 'Binance Coin (BNB)' }
  ];

  const paymentMethods = [
    { value: 'all', label: 'Все способы' },
    { value: 'sberbank', label: 'Сбербанк' },
    { value: 'tinkoff', label: 'Тинькофф' },
    { value: 'alfa', label: 'Альфа-Банк' },
    { value: 'qiwi', label: 'QIWI' },
    { value: 'yandex', label: 'ЮMoney' }
  ];

  return (
    <div className="bg-gray-800 rounded-lg p-6 border border-gray-700 mb-3">
      <div className="flex flex-col gap-4 sm:flex-row sm:flex-wrap sm:items-end">

        {/* Currency Filter */}
        <div className="sm:w-48">
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Криптовалюта
          </label>
          <select
            value={filters.currency}
            onChange={(e) => onFiltersChange({ ...filters, currency: e.target.value })}
            className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            {currencies.map((currency) => (
              <option key={currency.value} value={currency.value}>
                {currency.label}
              </option>
            ))}
          </select>
        </div>

        {/* Payment Method Filter */}
        <div className="sm:w-48">
          <label className="block text-sm font-medium text-gray-300 mb-1">
            Способ оплаты
          </label>
          <select
            value={filters.paymentMethod}
            onChange={(e) => onFiltersChange({ ...filters, paymentMethod: e.target.value })}
            className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            {paymentMethods.map((method) => (
              <option key={method.value} value={method.value}>
                {method.label}
              </option>
            ))}
          </select>
        </div>

        {/* Clear Filters Button */}
        <button
          onClick={() => onFiltersChange({ currency: 'all', paymentMethod: 'all', amount: '' })}
          className="sm:w-auto py-2 px-4 bg-gray-700 hover:bg-gray-600 text-gray-300 rounded-lg transition-colors"
        >
          Очистить фильтры
        </button>
      </div>

      {/* Tabs */}
      <div className="flex space-x-1 mt-4 bg-gray-700 rounded-lg p-1">
        <button
          onClick={() => onTabChange('buy')}
          className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
            activeTab === 'buy'
              ? 'bg-green-600 text-white'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Купить
        </button>
        <button
          onClick={() => onTabChange('sell')}
          className={`flex-1 py-2 px-4 rounded-md text-sm font-medium transition-colors ${
            activeTab === 'sell'
              ? 'bg-red-600 text-white'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Продать
        </button>
      </div>

      {/* Create Order Button */}
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
