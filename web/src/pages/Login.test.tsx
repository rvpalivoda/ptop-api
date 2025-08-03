import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { render, screen, fireEvent } from '@testing-library/react';
import Login from './Login';

const loginMock = vi.fn();
vi.mock('@/context/AuthContext', () => ({
  useAuth: () => ({ login: loginMock }),
}));

describe('Login page', () => {
  it('отправляет учетные данные при сабмите', async () => {
    loginMock.mockResolvedValue(undefined);
    render(
      <MemoryRouter>
        <Login />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByPlaceholderText(/username/i), {
      target: { value: 'user' },
    });
    fireEvent.change(screen.getByPlaceholderText(/Введите пароль/i), {
      target: { value: 'pass' },
    });
    fireEvent.click(screen.getByRole('button', { name: /войти/i }));

    await screen.findByText(/Забыли пароль\?/i);
    expect(loginMock).toHaveBeenCalledWith('user', 'pass');
  });
});
