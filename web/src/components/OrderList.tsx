
import { OrderCard } from './OrderCard';

interface OrderListProps {
  type: 'buy' | 'sell';
  filters: {
    fromAsset: string;
    toAsset: string;
    minAmount: string;
    maxAmount: string;
    paymentMethod: string;
  };
}

export const OrderList = ({ type, filters }: OrderListProps) => {
  // Mock data for demonstration
  const mockOrders = [
    {
      id: '1',
      trader: {
        name: 'CryptoTrader98',
        rating: 4.9,
        completedTrades: 1247,
        online: true
      },
      currency: 'BTC',
      amount: '0.5',
      price: '2,845,000',
      paymentMethods: ['Сбербанк', 'Тинькофф'],
      limits: { min: '50,000', max: '500,000' },
      type: type
    },
    {
      id: '2',
      trader: {
        name: 'BitMaster',
        rating: 4.8,
        completedTrades: 892,
        online: true
      },
      currency: 'ETH',
      amount: '2.5',
      price: '185,000',
      paymentMethods: ['Альфа-Банк', 'QIWI'],
      limits: { min: '20,000', max: '300,000' },
      type: type
    },
    {
      id: '3',
      trader: {
        name: 'CoinExpert',
        rating: 4.95,
        completedTrades: 2156,
        online: false
      },
      currency: 'USDT',
      amount: '10,000',
      price: '95.5',
      paymentMethods: ['ЮMoney', 'Сбербанк'],
      limits: { min: '5,000', max: '100,000' },
      type: type
    }
  ];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-white">
          {type === 'buy' ? 'Объявления о покупке' : 'Объявления о продаже'}
        </h3>
        <span className="text-sm text-gray-400">
          Найдено: {mockOrders.length} объявлений
        </span>
      </div>
      
      {mockOrders.map((order) => (
        <OrderCard key={order.id} order={order} />
      ))}
    </div>
  );
};
