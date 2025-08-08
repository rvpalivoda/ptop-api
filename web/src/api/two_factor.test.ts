import { describe, it, expect, vi, beforeEach } from "vitest";
import { enable2fa, disable2fa } from "./two_factor";

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

const mockFetch = vi.spyOn(global, "fetch" as any);

describe("two_factor api", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    localStorage.clear();
  });

  it("enable2fa отправляет пароль и получает секрет", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ secret: "sec", url: "otp" }),
    } as any);

    const res = await enable2fa("pass");
    expect(mockFetch.mock.calls[0][0]).toContain("/auth/2fa/enable");
    expect(mockFetch.mock.calls[0][1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ password: "pass" }),
    });
    expect(res).toEqual({ secret: "sec", url: "otp" });
  });

  it("disable2fa отправляет пароль", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: "2fa disabled" }),
    } as any);

    await disable2fa("pass");
    expect(mockFetch.mock.calls[0][0]).toContain("/auth/2fa/disable");
    expect(mockFetch.mock.calls[0][1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ password: "pass" }),
    });
  });
});
