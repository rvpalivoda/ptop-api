import {useState} from 'react';
import {Link, useNavigate} from 'react-router-dom';
import {Eye, EyeOff, ArrowLeft} from 'lucide-react';
import {toast} from '@/components/ui/sonner';

import {useAuth} from '@/context/AuthContext';

const Login = () => {
    const navigate = useNavigate();
    const {login} = useAuth();
    const [showPassword, setShowPassword] = useState(false);
    const [formData, setFormData] = useState({
        username: '',
        password: ''
    });

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        try {
            await login(formData.username, formData.password);
            toast('Вы успешно вошли в систему');
            navigate('/');
        } catch (err) {
            console.error('Login error:', err);
            toast(
                err instanceof Error ? `Ошибка входа: ${err.message}` : 'Ошибка входа',
            );
        }
    };

    return (
        <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center px-4">
            <div className="w-full max-w-md">
                <div className="bg-gray-800 rounded-lg shadow-sm border border-gray-700 p-8">
                    <div className="mb-8">
                        <Link to="/" className="inline-flex items-center text-gray-400 hover:text-white mb-4">
                            <ArrowLeft className="w-4 h-4 mr-2"/>
                            Назад
                        </Link>
                        <h1 className="text-2xl font-semibold mb-2">Вход в систему</h1>
                        <p className="text-gray-300">Войдите в свой аккаунт для доступа к торговле</p>
                    </div>

                    <form onSubmit={handleSubmit} className="space-y-6">
                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-2">
                                Имя пользователя
                            </label>
                            <input
                                type="text"
                                required
                                value={formData.username}
                                onChange={(e) => setFormData({...formData, username: e.target.value})}
                                className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                placeholder="username"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-300 mb-2">
                                Пароль
                            </label>
                            <div className="relative">
                                <input
                                    type={showPassword ? 'text' : 'password'}
                                    required
                                    value={formData.password}
                                    onChange={(e) => setFormData({...formData, password: e.target.value})}
                                    className="w-full px-3 py-2 pr-10 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    placeholder="Введите пароль"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword(!showPassword)}
                                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200"
                                >
                                    {showPassword ? <EyeOff className="w-4 h-4"/> : <Eye className="w-4 h-4"/>}
                                </button>
                            </div>
                        </div>


                        <button
                            type="submit"
                            className="w-full bg-blue-600 text-white py-2  rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                        >
                            Войти
                        </button>
                    </form>

                    <div className="mt-10 flex justify-between text-sm text-gray-300">
                        <Link to="/register" className="text-blue-400 hover:text-blue-300 font-medium">
                            Зарегистрироваться
                        </Link>
                        <Link to="/recover" className="text-blue-400 hover:text-blue-300 font-medium">
                            Забыли пароль?
                        </Link>
                    </div>

                </div>
            </div>
        </div>
    );
};

export default Login;
