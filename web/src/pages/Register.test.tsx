import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { render, screen, fireEvent } from '@testing-library/react';
import Register from './Register';

const registerMock = vi.fn();
vi.mock('@/context', () => ({ useAuth: () => ({ register: registerMock }) }));

const words = Array.from({ length: 12 }, (_, i) => ({ position: i + 1, word: `w${i + 1}` }));

// Mock clipboard
Object.assign(navigator, {
  clipboard: { writeText: vi.fn() },
});

describe('Register page', () => {
  it('показывает мнемонику после успешной регистрации', async () => {
    registerMock.mockResolvedValue({ mnemonic: words });
    render(
      <MemoryRouter>
        <Register />
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByPlaceholderText(/Username/i), {
      target: { value: 'user' },
    });
    fireEvent.change(screen.getByPlaceholderText(/Минимум 8 символов/i), {
      target: { value: 'password1' },
    });
    fireEvent.change(screen.getByPlaceholderText(/Повторите пароль/i), {
      target: { value: 'password1' },
    });

    fireEvent.click(screen.getByRole('button', { name: /Создать аккаунт/i }));

    expect(await screen.findByText('1. w1')).toBeInTheDocument();
    expect(registerMock).toHaveBeenCalledWith('user', 'password1', 'password1');
  });
});
