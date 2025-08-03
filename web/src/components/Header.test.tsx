import { describe, it, expect, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { render, screen } from '@testing-library/react';
import { Header } from './Header';

const authState = { isAuthenticated: false };
vi.mock('@/context', () => ({ useAuth: () => authState }));
vi.mock('./ProfileDrawer', () => ({ ProfileDrawer: () => <div>drawer</div> }));

describe('Header component', () => {
  it('рендерит ссылки входа и регистрации для гостя', () => {
    authState.isAuthenticated = false;
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>,
    );
    expect(screen.getByText('Вход')).toBeInTheDocument();
    expect(screen.getByText('Регистрация')).toBeInTheDocument();
  });

  it('рендерит профиль для авторизованного пользователя', () => {
    authState.isAuthenticated = true;
    render(
      <MemoryRouter>
        <Header />
      </MemoryRouter>,
    );
    expect(screen.getByText('drawer')).toBeInTheDocument();
  });
});
