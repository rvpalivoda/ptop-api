import { createContext, useContext, useEffect, useState } from "react";
import {
  login as apiLogin,
  register as apiRegister,
  logout as apiLogout,
  refresh as apiRefresh,
  recover as apiRecover,
  regenerateWords as apiRegenerate,
  changePassword as apiChangePassword,
  type RegisterResponse,
} from "@/api/auth";
import { setPinCode as apiSetPinCode } from "@/api/pin";
import { disable2fa as apiDisable2fa } from "@/api/two_factor";
import {
  loadTokens,
  saveTokens,
  clearTokens,
  type Tokens,
} from "@/storage/token";
import {
  loadUserInfo,
  saveUserInfo,
  clearUserInfo,
  type UserInfo,
} from "@/storage/user";

interface AuthContextValue {
  tokens: Tokens | null;
  isAuthenticated: boolean;
  userInfo: UserInfo | null;
  login: (username: string, password: string, captcha: string) => Promise<void>;
  register: (
    username: string,
    password: string,
    captcha: string,
  ) => Promise<RegisterResponse>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  recover: (
    username: string,
    words: string[],
    indices: number[],
    captcha: string,
    newPassword?: string,
  ) => Promise<void>;
  regenerateWords: (password: string) => Promise<RegisterResponse>;
  changePassword: (current: string, newPwd: string) => Promise<void>;
  disable2fa: (password: string) => Promise<void>;
  setPinCode: (password: string, pin: string) => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [tokens, setTokens] = useState<Tokens | null>(loadTokens());
  const [userInfo, setUserInfo] = useState<UserInfo | null>(loadUserInfo());

  useEffect(() => {
    const t = loadTokens();
    setTokens(t);
  }, []);

  const login = async (
    username: string,
    password: string,
    captcha: string,
  ): Promise<void> => {
    const t = await apiLogin(username, password, captcha);
    saveTokens(t);
    setTokens(t);
    const payload = JSON.parse(atob(t.access.split(".")[1] || ""));
    const info: UserInfo = { username, name: payload.sub || "" };
    saveUserInfo(info);
    setUserInfo(info);
  };

  const register = async (
    username: string,
    password: string,
    captcha: string,
  ): Promise<RegisterResponse> => {
    const res = await apiRegister(username, password, captcha);
    return res;
  };

  const logout = async (): Promise<void> => {
    await apiLogout();
    clearTokens();
    setTokens(null);
    clearUserInfo();
    setUserInfo(null);
  };

  const refresh = async (): Promise<void> => {
    const t = await apiRefresh();
    saveTokens(t);
    setTokens(t);
    setUserInfo((prev) => {
      const payload = JSON.parse(atob(t.access.split(".")[1] || ""));
      const info = { username: prev?.username || "", name: payload.sub || "" };
      saveUserInfo(info);
      return info;
    });
  };

  const recover = async (
    username: string,
    words: string[],
    indices: number[],
    captcha: string,
    newPassword?: string,
  ): Promise<void> => {
    const t = await apiRecover(username, words, indices, captcha, newPassword);
    saveTokens(t);
    setTokens(t);
    const payload = JSON.parse(atob(t.access.split(".")[1] || ""));
    const info: UserInfo = { username, name: payload.sub || "" };
    saveUserInfo(info);
    setUserInfo(info);
  };

  const regenerateWords = async (password: string): Promise<RegisterResponse> => {
    return apiRegenerate(password);
  };

  const changePassword = async (
    current: string,
    newPwd: string,
  ): Promise<void> => {
    await apiChangePassword(current, newPwd);
  };

  const disable2fa = async (password: string): Promise<void> => {
    await apiDisable2fa(password);
    // optionally refresh state
  };

  const setPinCode = async (
    password: string,
    pin: string,
  ): Promise<void> => {
    await apiSetPinCode(password, pin);
  };

  return (
    <AuthContext.Provider
      value={{
        tokens,
        isAuthenticated: !!tokens,
        userInfo,
        login,
        register,
        logout,
        refresh,
        recover,
        regenerateWords,
        changePassword,
        disable2fa,
        setPinCode,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextValue => {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
};
