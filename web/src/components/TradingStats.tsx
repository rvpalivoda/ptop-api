
import { TrendingUp, TrendingDown, DollarSign, Users } from 'lucide-react';

export const TradingStats = () => {
  const stats = [
    {
      title: 'BTC/RUB',
      value: '₽2,845,000',
      change: '+2.5%',
      isPositive: true,
      icon: TrendingUp
    },
    {
      title: 'ETH/RUB',
      value: '₽185,000',
      change: '-1.2%',
      isPositive: false,
      icon: TrendingDown
    },
    {
      title: 'Активные сделки',
      value: '1,247',
      change: '+15%',
      isPositive: true,
      icon: DollarSign
    },
    {
      title: 'Онлайн трейдеров',
      value: '3,892',
      change: '+8%',
      isPositive: true,
      icon: Users
    }
  ];

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      {stats.map((stat, index) => {
        const Icon = stat.icon;
        return (
          <div key={index} className="bg-gray-800 rounded-lg p-6 border border-gray-700">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-gray-400 text-sm">{stat.title}</p>
                <p className="text-2xl font-bold text-white mt-1">{stat.value}</p>
                <p className={`text-sm mt-1 ${stat.isPositive ? 'text-green-400' : 'text-red-400'}`}>
                  {stat.change}
                </p>
              </div>
              <div className={`p-3 rounded-full ${stat.isPositive ? 'bg-green-600/20' : 'bg-red-600/20'}`}>
                <Icon className={`h-6 w-6 ${stat.isPositive ? 'text-green-400' : 'text-red-400'}`} />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
