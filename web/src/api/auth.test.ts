import { describe, it, expect, vi, beforeEach } from "vitest";
/* eslint-disable @typescript-eslint/no-explicit-any */
import {
  login,
  register,
  recover,
  refresh,
  profile,
  recoverChallenge,
  changePassword,
  verifyPassword,
  logout,
} from "./auth";

const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: (key: string) => store[key] || null,
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    clear: () => {
      store = {};
    },
  };
})();

Object.defineProperty(global, "localStorage", { value: localStorageMock });

beforeEach(() => {
  vi.resetAllMocks();
  localStorage.clear();
});

describe("auth api", () => {
  it("login отправляет параметры и возвращает токены", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ access_token: "a", refresh_token: "r" }),
      } as any);

    const tokens = await login("user", "pass");
    expect(mockFetch).toHaveBeenCalled();
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({ username: "user", password: "pass" });
    expect(tokens).toEqual({ access: "a", refresh: "r" });
  });

  it("register отправляет параметры и возвращает мнемонику", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          access_token: "a",
          refresh_token: "r",
          mnemonic: [
            { position: 1, word: "one" },
            { position: 2, word: "two" },
          ],
        }),
      } as any);

    const res = await register("user", "pass", "pass");
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({
      username: "user",
      password: "pass",
      password_confirm: "pass",
    });
    expect(res).toEqual({
      access: "a",
      refresh: "r",
      mnemonic: [
        { position: 1, word: "one" },
        { position: 2, word: "two" },
      ],
    });
  });

  it("recover отправляет правильный payload", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ access_token: "a", refresh_token: "r" }),
      } as any);

    const phrases = [
      { position: 1, word: "w1" },
      { position: 2, word: "w2" },
      { position: 3, word: "w3" },
    ];
    const res = await recover("user", phrases, "new", "new");
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({
      username: "user",
      phrases,
      new_password: "new",
      password_confirm: "new",
    });
    expect(res).toEqual({ access: "a", refresh: "r" });
  });

  it("recoverChallenge возвращает позиции", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ positions: [1, 2, 3] }),
      } as any);

    const res = await recoverChallenge("user");
    expect(mockFetch).toHaveBeenCalled();
    expect(res).toEqual({ positions: [1, 2, 3] });
  });

  it("changePassword отправляет корректный payload", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ status: "ok" }),
      } as any);

    await changePassword("old", "new");
    const [url, opts] = mockFetch.mock.calls[0];
    expect((url as string).endsWith("/auth/password")).toBe(true);
    const body = JSON.parse(opts.body);
    expect(body).toEqual({
      old_password: "old",
      new_password: "new",
      confirm_password: "new",
    });
  });

  it("verifyPassword отправляет пароль и возвращает результат", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ verified: true }),
      } as any);

    const res = await verifyPassword("pwd");
    const [url, opts] = mockFetch.mock.calls[0];
    expect((url as string).endsWith("/auth/verify-password")).toBe(true);
    const body = JSON.parse(opts.body);
    expect(body).toEqual({ password: "pwd" });
    expect(res).toEqual({ verified: true });
  });

  it("logout делает POST без тела", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ status: "ok" }),
      } as any);

    await logout();
    const [url, opts] = mockFetch.mock.calls[0];
    expect((url as string).endsWith("/auth/logout")).toBe(true);
    expect(opts.method).toBe("POST");
    expect(opts.body).toBeUndefined();
  });

  it("refresh использует refresh_token", async () => {
    localStorage.setItem("peerex_tokens", JSON.stringify({ access: "x", refresh: "r" }));
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ access_token: "na", refresh_token: "nr" }),
      } as any);

    const res = await refresh();
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({ refresh_token: "r" });
    expect(res).toEqual({ access: "na", refresh: "nr" });
  });

  it("profile возвращает данные профиля", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({
          username: "u",
          twofa_enabled: true,
          pincode_set: false,
        }),
      } as any);

    const res = await profile();
    expect(mockFetch).toHaveBeenCalled();
    expect(res).toEqual({
      username: "u",
      twofa_enabled: true,
      pincode_set: false,
    });
  });
});

