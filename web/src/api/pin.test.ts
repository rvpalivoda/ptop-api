import { describe, it, expect, vi, beforeEach } from "vitest";
import { setPinCode } from "./pin";

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

describe("pin api", () => {
  beforeEach(() => {
    vi.resetAllMocks();
    localStorage.clear();
  });

  it("setPinCode отправляет пароль и пинкод", async () => {
    mockFetch.mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ status: "ok" }),
    } as any);

    const res = await setPinCode("pass", "1234");
    expect(mockFetch.mock.calls[0][0]).toContain("/auth/pincode");
    expect(mockFetch.mock.calls[0][1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ password: "pass", pincode: "1234" }),
    });
    expect(res).toEqual({ status: "ok" });
  });
});

