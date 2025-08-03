
import { Star, Circle, MessageCircle } from 'lucide-react';

interface OrderCardProps {
  order: {
    id: string;
    trader: {
      name: string;
      rating: number;
      completedTrades: number;
      online: boolean;
    };
    currency: string;
    amount: string;
    price: string;
    paymentMethods: string[];
    limits: { min: string; max: string };
    type: 'buy' | 'sell';
  };
}

export const OrderCard = ({ order }: OrderCardProps) => {
  const { trader, currency, amount, price, paymentMethods, limits, type } = order;

  return (
    <div className="bg-gray-700 rounded-lg p-6 border border-gray-600 hover:border-gray-500 transition-colors">
      <div className="flex flex-col lg:flex-row lg:items-center lg:justify-between space-y-4 lg:space-y-0">
        {/* Trader Info */}
        <div className="flex items-center space-x-4">
          <div className="relative">
            <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center text-white font-bold">
              {trader.name.charAt(0)}
            </div>
            <div className={`absolute -bottom-1 -right-1 w-4 h-4 rounded-full border-2 border-gray-700 ${
              trader.online ? 'bg-green-500' : 'bg-gray-500'
            }`}>
              <Circle className="w-full h-full" />
            </div>
          </div>
          <div>
            <h4 className="font-semibold text-white">{trader.name}</h4>
            <div className="flex items-center space-x-2 text-sm text-gray-400">
              <div className="flex items-center space-x-1">
                <Star className="w-4 h-4 text-yellow-400 fill-current" />
                <span>{trader.rating}</span>
              </div>
              <span>•</span>
              <span>{trader.completedTrades} сделок</span>
            </div>
          </div>
        </div>

        {/* Order Details */}
        <div className="flex-1 lg:mx-8">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <div>
              <p className="text-sm text-gray-400">Валюта</p>
              <p className="font-semibold text-white">{currency}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">Количество</p>
              <p className="font-semibold text-white">{amount}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">Цена</p>
              <p className="font-semibold text-white">₽{price}</p>
            </div>
            <div>
              <p className="text-sm text-gray-400">Лимиты</p>
              <p className="font-semibold text-white">₽{limits.min} - ₽{limits.max}</p>
            </div>
          </div>
          
          <div className="mt-3">
            <p className="text-sm text-gray-400 mb-2">Способы оплаты:</p>
            <div className="flex flex-wrap gap-2">
              {paymentMethods.map((method, index) => (
                <span
                  key={index}
                  className="px-2 py-1 bg-gray-600 text-gray-300 text-xs rounded-full"
                >
                  {method}
                </span>
              ))}
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex space-x-3">
          <button className="flex items-center space-x-2 px-4 py-2 bg-gray-600 hover:bg-gray-500 text-white rounded-lg transition-colors">
            <MessageCircle className="w-4 h-4" />
            <span>Чат</span>
          </button>
          <button className={`px-6 py-2 rounded-lg font-medium transition-colors ${
            type === 'buy'
              ? 'bg-green-600 hover:bg-green-700 text-white'
              : 'bg-red-600 hover:bg-red-700 text-white'
          }`}>
            {type === 'buy' ? 'Продать' : 'Купить'}
          </button>
        </div>
      </div>
    </div>
  );
};
