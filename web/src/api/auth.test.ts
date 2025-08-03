import { describe, it, expect, vi, beforeEach } from "vitest";
import { login, register, recover, refresh } from "./auth";

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
          mnemonic: "one two",
        }),
      } as any);

    const res = await register("user", "pass", "pass");
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({
      username: "user",
      password: "pass",
      password_confirm: "pass",
    });
    expect(res).toEqual({ access: "a", refresh: "r", mnemonic: "one two" });
  });

  it("recover отправляет правильный payload", async () => {
    const mockFetch = vi
      .spyOn(global, "fetch" as any)
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: async () => ({ access_token: "a", refresh_token: "r" }),
      } as any);

    const res = await recover(
      "user",
      ["w1", "w2", "w3"],
      [1, 2, 3],
      "new",
      "new",
    );
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toEqual({
      username: "user",
      words: ["w1", "w2", "w3"],
      indices: [1, 2, 3],
      new_password: "new",
      password_confirm: "new",
    });
    expect(res).toEqual({ access: "a", refresh: "r" });
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
});

