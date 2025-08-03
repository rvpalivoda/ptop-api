import {useState} from 'react';
import Captcha from '@/components/Captcha';
import {Link} from 'react-router-dom';
import {Eye, EyeOff, ArrowLeft, Copy, Download} from 'lucide-react';
import {toast} from '@/components/ui/sonner';
import {useAuth} from '@/context';

const Register = () => {
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const {register} = useAuth();
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        confirmPassword: '',
    });
    const [captcha, setCaptcha] = useState('');
    const [seed, setSeed] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();

        if (formData.password.length < 8) {
            toast('Пароль слишком короткий. Минимум 8 символов');
            return;
        }
        if (formData.password !== formData.confirmPassword) {
            toast('Пароли не совпадают');
            return;
        }

        try {
            const {seed} = await register(
                formData.username,
                formData.password,
                captcha,
            );
            setSeed(seed);
            toast('Успешная регистрация');
        } catch (err) {
            console.error('Registration error:', err);
            toast(
                err instanceof Error
                    ? `Ошибка регистрации: ${err.message}`
                    : 'Не удалось завершить регистрацию',
            );
        }
    };

    return (
        <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center px-4 py-8">
            <div className="w-full max-w-md">
                <div className="bg-gray-800 rounded-lg shadow-sm border border-gray-700 p-8">
                    <div className="mb-8">
                        <Link
                            to="/"
                            className="inline-flex items-center text-gray-400 hover:text-white mb-4"
                        >
                            <ArrowLeft className="w-4 h-4 mr-2"/>
                            Назад
                        </Link>
                        <h1 className="text-2xl font-semibold mb-2">Регистрация</h1>
                        <p className="text-gray-300">Создайте аккаунт для начала торговли</p>
                    </div>

                    {!seed ? (
                        <form onSubmit={handleSubmit} className="space-y-6">
                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Имя пользователя
                                </label>
                                <input
                                    type="text"
                                    required
                                    value={formData.username}
                                    onChange={(e) =>
                                        setFormData({...formData, username: e.target.value})
                                    }
                                    className="w-full px-3 py-2 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                    placeholder="Username"
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
                                        onChange={(e) =>
                                            setFormData({...formData, password: e.target.value})
                                        }
                                        className="w-full px-3 py-2 pr-10 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        placeholder="Минимум 8 символов"
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

                            <div>
                                <label className="block text-sm font-medium text-gray-300 mb-2">
                                    Подтвердите пароль
                                </label>
                                <div className="relative">
                                    <input
                                        type={showConfirmPassword ? 'text' : 'password'}
                                        required
                                        value={formData.confirmPassword}
                                        onChange={(e) =>
                                            setFormData({...formData, confirmPassword: e.target.value})
                                        }
                                        className="w-full px-3 py-2 pr-10 bg-gray-700 border border-gray-600 rounded-md text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                        placeholder="Повторите пароль"
                                    />
                                    <button
                                        type="button"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                        className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-200"
                                    >
                                        {showConfirmPassword ? <EyeOff className="w-4 h-4"/> :
                                            <Eye className="w-4 h-4"/>}
                                    </button>
                                </div>
                            </div>


                            <button
                                type="submit"
                                className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                            >
                                Создать аккаунт
                            </button>
                            <Captcha onChange={setCaptcha}/>
                        </form>
                    ) : (
                        <div className="space-y-4">
                            <div className="p-4 bg-gray-700 border border-gray-600 rounded-md text-center">
                                <p className="mb-2 text-sm text-gray-300">Ваша seed-фраза:</p>
                                <div className="grid grid-cols-2 gap-2 mb-4">
                                    {seed.split(' ').map((word, idx) => (
                                        <div
                                            key={idx}
                                            className="px-3 py-1 bg-gray-800 border border-gray-600 rounded-md text-sm"
                                        >
                                            {idx + 1}. {word}
                                        </div>
                                    ))}
                                </div>
                                <div className="flex items-center justify-center space-x-4">
                                    <button
                                        type="button"
                                        onClick={() => navigator.clipboard.writeText(seed)}
                                        className="flex items-center text-blue-400 hover:text-blue-300"
                                    >
                                        <Copy className="w-4 h-4 mr-1"/> Скопировать
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            const blob = new Blob([JSON.stringify(seed.split(' '))], {type: 'application/json'});
                                            const url = URL.createObjectURL(blob);
                                            const link = document.createElement('a');
                                            link.href = url;
                                            link.download = 'words.json';
                                            link.click();
                                            URL.revokeObjectURL(url);
                                        }}
                                        className="flex items-center text-blue-400 hover:text-blue-300"
                                    >
                                        <Download className="w-4 h-4 mr-1"/> Скачать
                                    </button>
                                </div>
                            </div>
                            <div className="text-center">
                                <Link to="/login" className="text-blue-400 hover:text-blue-300 font-medium">
                                    Перейти к входу
                                </Link>
                            </div>
                        </div>
                    )}


                    <div className="mt-10 flex justify-between text-sm text-gray-300">
                        <Link to="/login" className="text-blue-400 hover:text-blue-300 font-medium">
                            Войти
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

export default Register;
