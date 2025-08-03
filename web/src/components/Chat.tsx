
import { useState } from 'react';
import { Send, Search, User, MoreVertical } from 'lucide-react';

interface Message {
  id: number;
  text: string;
  sender: 'user' | 'other';
  timestamp: string;
  senderName: string;
}

interface ChatUser {
  id: number;
  name: string;
  lastMessage: string;
  timestamp: string;
  unread: number;
  avatar?: string;
}

export const Chat = () => {
  const [selectedChat, setSelectedChat] = useState<number | null>(1);
  const [messageText, setMessageText] = useState('');
  const [searchQuery, setSearchQuery] = useState('');

  const chatUsers: ChatUser[] = [
    {
      id: 1,
      name: 'Алексей Морозов',
      lastMessage: 'Готов к сделке по USDT',
      timestamp: '10:30',
      unread: 2
    },
    {
      id: 2,
      name: 'Мария Сидорова',
      lastMessage: 'Спасибо за быструю сделку!',
      timestamp: '09:15',
      unread: 0
    },
    {
      id: 3,
      name: 'Дмитрий Козлов',
      lastMessage: 'Когда можем встретиться?',
      timestamp: 'Вчера',
      unread: 1
    },
    {
      id: 4,
      name: 'Елена Петрова',
      lastMessage: 'Подтверждаю получение',
      timestamp: 'Вчера',
      unread: 0
    }
  ];

  const messages: Message[] = [
    {
      id: 1,
      text: 'Здравствуйте! Интересует покупка USDT',
      sender: 'other',
      timestamp: '10:15',
      senderName: 'Алексей Морозов'
    },
    {
      id: 2,
      text: 'Здравствуйте! Да, есть в наличии. Какая сумма вас интересует?',
      sender: 'user',
      timestamp: '10:16',
      senderName: 'Вы'
    },
    {
      id: 3,
      text: 'Нужно на 50,000 рублей. Курс какой?',
      sender: 'other',
      timestamp: '10:18',
      senderName: 'Алексей Морозов'
    },
    {
      id: 4,
      text: 'Курс 94.5 рубля за USDT. Способ оплаты?',
      sender: 'user',
      timestamp: '10:20',
      senderName: 'Вы'
    },
    {
      id: 5,
      text: 'Сбербанк подходит. Готов к сделке по USDT',
      sender: 'other',
      timestamp: '10:30',
      senderName: 'Алексей Морозов'
    }
  ];

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (messageText.trim()) {
      console.log('Sending message:', messageText);
      setMessageText('');
    }
  };

  const filteredUsers = chatUsers.filter(user =>
    user.name.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const selectedUser = chatUsers.find(user => user.id === selectedChat);

  return (
    <div className="flex h-96 border border-gray-700 rounded-lg overflow-hidden bg-gray-800 text-white">
      {/* Chat List */}
      <div className="w-1/3 bg-gray-800 border-r border-gray-700">
        <div className="p-4 border-b border-gray-700">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              placeholder="Поиск чатов..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-10 pr-4 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
            />
          </div>
        </div>

        <div className="overflow-y-auto">
          {filteredUsers.map((user) => (
            <div
              key={user.id}
              onClick={() => setSelectedChat(user.id)}
              className={`p-4 border-b border-gray-700 cursor-pointer hover:bg-gray-700 transition-colors ${
                selectedChat === user.id ? 'bg-gray-700 border-r-2 border-blue-600' : ''
              }`}
            >
              <div className="flex items-start space-x-3">
                <div className="w-10 h-10 bg-gray-700 rounded-full flex items-center justify-center">
                  <User className="w-5 h-5 text-gray-300" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <h4 className="text-sm font-medium truncate">{user.name}</h4>
                    <span className="text-xs text-gray-400">{user.timestamp}</span>
                  </div>
                  <div className="flex items-center justify-between mt-1">
                    <p className="text-sm text-gray-400 truncate">{user.lastMessage}</p>
                    {user.unread > 0 && (
                      <span className="bg-blue-600 text-white text-xs rounded-full px-2 py-1 min-w-[20px] text-center">
                        {user.unread}
                      </span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Chat Content */}
      <div className="flex-1 flex flex-col">
        {selectedUser ? (
          <>
            {/* Chat Header */}
            <div className="p-4 border-b border-gray-700 bg-gray-800">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-8 h-8 bg-gray-700 rounded-full flex items-center justify-center">
                    <User className="w-4 h-4 text-gray-300" />
                  </div>
                  <div>
                    <h3 className="text-sm font-medium">{selectedUser.name}</h3>
                    <p className="text-xs text-green-600">В сети</p>
                  </div>
                </div>
                <button className="text-gray-400 hover:text-gray-200">
                  <MoreVertical className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-900">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className={`flex ${message.sender === 'user' ? 'justify-end' : 'justify-start'}`}
                >
                  <div
                    className={`max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${
                      message.sender === 'user'
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-700 text-white border border-gray-600'
                    }`}
                  >
                    <p className="text-sm">{message.text}</p>
                    <p
                      className={`text-xs mt-1 ${
                        message.sender === 'user' ? 'text-blue-100' : 'text-gray-400'
                      }`}
                    >
                      {message.timestamp}
                    </p>
                  </div>
                </div>
              ))}
            </div>

            {/* Message Input */}
            <div className="p-4 border-t border-gray-700 bg-gray-800">
              <form onSubmit={handleSendMessage} className="flex space-x-2">
                <input
                  type="text"
                  value={messageText}
                  onChange={(e) => setMessageText(e.target.value)}
                  placeholder="Введите сообщение..."
                  className="flex-1 px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent text-sm"
                />
                <button
                  type="submit"
                  className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                >
                  <Send className="w-4 h-4" />
                </button>
              </form>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center bg-gray-900">
            <div className="text-center">
              <div className="w-16 h-16 bg-gray-700 rounded-full flex items-center justify-center mx-auto mb-4">
                <User className="w-8 h-8 text-gray-400" />
              </div>
              <h3 className="text-lg font-medium mb-2">Выберите чат</h3>
              <p className="text-gray-400">Выберите собеседника для начала общения</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
