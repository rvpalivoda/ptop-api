/* @vitest-environment jsdom */
import React from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { createRoot } from "react-dom/client";
import { act } from "react";
import { AuthProvider, useAuth } from "./AuthContext";

(globalThis as unknown as { IS_REACT_ACT_ENVIRONMENT: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

vi.mock("@/api/auth", () => ({
  login: vi.fn(),
  register: vi.fn(),
  refresh: vi.fn(),
  recover: vi.fn(),
  regenerateWords: vi.fn(),
  changePassword: vi.fn(),
  profile: vi.fn(),
  logout: vi.fn(),
}));
vi.mock("@/api/pin", () => ({ setPinCode: vi.fn() }));

describe("AuthContext", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  it("login получает профиль и сохраняет userInfo", async () => {
    const { login, profile } = await import("@/api/auth");
    vi.mocked(login).mockResolvedValue({ access: "a", refresh: "r" });
    vi.mocked(profile).mockResolvedValue({
      username: "user",
      twofa_enabled: true,
      pincode_set: false,
    });

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };

    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    await act(async () => {
      await ctx!.login("user", "pass");
    });

    expect(login).toHaveBeenCalledWith("user", "pass");
    expect(profile).toHaveBeenCalled();
    expect(ctx!.userInfo).toEqual({
      username: "user",
      twofaEnabled: true,
      pinCodeSet: false,
    });
  });

  it("register возвращает мнемонику", async () => {
    const { register } = await import("@/api/auth");
    vi.mocked(register).mockResolvedValue({
      access: "a",
      refresh: "r",
      mnemonic: [
        { position: 1, word: "one" },
        { position: 2, word: "two" },
      ],
    });

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };

    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    let res: any;
    await act(async () => {
      res = await ctx!.register("u", "p", "p");
    });

    expect(register).toHaveBeenCalledWith("u", "p", "p");
    expect(res).toEqual({
      access: "a",
      refresh: "r",
      mnemonic: [
        { position: 1, word: "one" },
        { position: 2, word: "two" },
      ],
    });
  });

  it("refresh обновляет токены и userInfo", async () => {
    const { refresh, profile } = await import("@/api/auth");
    vi.mocked(refresh).mockResolvedValue({ access: "na", refresh: "nr" });
    vi.mocked(profile).mockResolvedValue({
      username: "u2",
      twofa_enabled: false,
      pincode_set: true,
    });

    localStorage.setItem("peerex_tokens", JSON.stringify({ access: "a", refresh: "r" }));

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };
    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    await act(async () => {
      await ctx!.refresh();
    });

    expect(refresh).toHaveBeenCalled();
    expect(ctx!.userInfo).toEqual({ username: "u2", twofaEnabled: false, pinCodeSet: true });
  });

  it("recover устанавливает userInfo", async () => {
    const { recover, profile } = await import("@/api/auth");
    vi.mocked(recover).mockResolvedValue({ access: "a", refresh: "r" });
    vi.mocked(profile).mockResolvedValue({
      username: "u3",
      twofa_enabled: true,
      pincode_set: true,
    });

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };
    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    await act(async () => {
      await ctx!.recover("u", [], "n", "n");
    });

    expect(recover).toHaveBeenCalledWith("u", [], "n", "n");
    expect(ctx!.userInfo).toEqual({ username: "u3", twofaEnabled: true, pinCodeSet: true });
  });

  it("setPinCode вызывает API", async () => {
    const { setPinCode } = await import("@/api/pin");
    vi.mocked(setPinCode).mockResolvedValue(undefined);

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };
    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    await act(async () => {
      await ctx!.setPinCode("pwd", "1234");
    });

    expect(setPinCode).toHaveBeenCalledWith("pwd", "1234");
  });

  it("logout вызывает API и очищает состояние", async () => {
    const { logout: apiLogout } = await import("@/api/auth");
    vi.mocked(apiLogout).mockResolvedValue({ status: "ok" } as any);

    localStorage.setItem(
      "peerex_tokens",
      JSON.stringify({ access: "a", refresh: "r" }),
    );
    localStorage.setItem(
      "peerex_user_info",
      JSON.stringify({ username: "u", twofaEnabled: false, pinCodeSet: false }),
    );

    let ctx: ReturnType<typeof useAuth> | undefined;
    const Consumer = () => {
      ctx = useAuth();
      return null;
    };
    const container = document.createElement("div");
    const root = createRoot(container);
    act(() => {
      root.render(
        <AuthProvider>
          <Consumer />
        </AuthProvider>,
      );
    });

    await act(async () => {
      await ctx!.logout();
    });

    expect(apiLogout).toHaveBeenCalled();
    expect(ctx!.tokens).toBeNull();
    expect(ctx!.userInfo).toBeNull();
    expect(localStorage.getItem("peerex_tokens")).toBeNull();
    expect(localStorage.getItem("peerex_user_info")).toBeNull();
  });
});
