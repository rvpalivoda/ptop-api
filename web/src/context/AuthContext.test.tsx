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
}));
vi.mock("@/api/pin", () => ({ setPinCode: vi.fn() }));
vi.mock("@/api/two_factor", () => ({ disable2fa: vi.fn() }));

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
});
