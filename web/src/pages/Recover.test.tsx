import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import Recover from './Recover';

const recoverMock = vi.fn();
vi.mock('@/context', () => ({ useAuth: () => ({ recover: recoverMock }) }));
vi.mock('@/api/auth', () => ({ recoverChallenge: vi.fn() }));
vi.mock('@/components/ui/sonner', () => ({ toast: vi.fn() }));

import { recoverChallenge } from '@/api/auth';
const recoverChallengeMock = vi.mocked(recoverChallenge);
recoverChallengeMock.mockResolvedValue({ positions: [1, 2, 3] });

describe('Recover page', () => {
  it('отправляет слова и новый пароль', async () => {
    recoverMock.mockResolvedValue(undefined);
    render(
      <MemoryRouter>
        <Recover />
      </MemoryRouter>,
    );

    const userField = screen.getByPlaceholderText(/username/i);
    fireEvent.change(userField, { target: { value: 'user' } });
    fireEvent.blur(userField);
    await waitFor(() => expect(recoverChallengeMock).toHaveBeenCalled());

    const textInputs = screen.getAllByRole('textbox');
    fireEvent.change(textInputs[1], { target: { value: 'one' } });
    fireEvent.change(textInputs[2], { target: { value: 'two' } });
    fireEvent.change(textInputs[3], { target: { value: 'three' } });
    await waitFor(() => document.querySelectorAll('input[type="password"]').length === 2);
    const passFields = document.querySelectorAll('input[type="password"]');
    fireEvent.change(passFields[0], { target: { value: 'password1' } });
    fireEvent.change(passFields[1], { target: { value: 'password1' } });

    fireEvent.click(screen.getByRole('button', { name: /Восстановить/i }));

    await waitFor(() =>
      expect(recoverMock).toHaveBeenCalledWith(
        'user',
        [
          { position: 1, word: 'one' },
          { position: 2, word: 'two' },
          { position: 3, word: 'three' },
        ],
        'password1',
        'password1',
      ),
    );
  });
});
