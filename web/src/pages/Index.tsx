
import { useState } from 'react';
import { Header } from '@/components/Header';
import { OrderList } from '@/components/OrderList';
import { CreateAdvertForm } from '@/components/CreateAdvertForm';
import { FilterPanel } from '@/components/FilterPanel';
//import { TradingStats } from '@/components/TradingStats';

const Index = () => {
  const [activeTab, setActiveTab] = useState<'buy' | 'sell'>('buy');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [filters, setFilters] = useState({
    fromAsset: 'all',
    toAsset: 'all',
    minAmount: '',
    maxAmount: '',
    paymentMethod: 'all'
  });

  return (
    <div className="min-h-screen bg-gray-900 text-white">
      <Header />
      
      <div className="container mx-auto px-4 py-0">
        {/* Trading
        <TradingStats />
        Stats */}
        {/* Main Content */}
        <div className="mt-8">
          <div className="bg-gray-800 rounded-lg p-6">
            <FilterPanel
              filters={filters}
              onFiltersChange={setFilters}
              activeTab={activeTab}
              onTabChange={setActiveTab}
              onCreate={() => setShowCreateForm(true)}
            />

            {/* Orders List */}
            <OrderList type={activeTab} filters={filters} />
            </div>
          </div>
        </div>
        
        {/* Create Advert Modal */}
        {showCreateForm && (
          <CreateAdvertForm
            onClose={() => setShowCreateForm(false)}
            type={activeTab}
          />
        )}
      </div>
  );
};

export default Index;
