import { createContext, useContext, useEffect, useState } from "react";
import {
  login as apiLogin,
  register as apiRegister,
  refresh as apiRefresh,
  recover as apiRecover,
  regenerateWords as apiRegenerate,
  changePassword as apiChangePassword,
  profile as apiProfile,
  logout as apiLogout,
  type RegisterResponse,
  type MnemonicResponse,
  type RecoverPhrase,
} from "@/api/auth";
import { setPinCode as apiSetPinCode } from "@/api/pin";
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
  login: (username: string, password: string) => Promise<void>;
  register: (
    username: string,
    password: string,
    passwordConfirm: string,
  ) => Promise<RegisterResponse>;
  logout: () => Promise<void>;
  refresh: () => Promise<void>;
  recover: (
    username: string,
    phrases: RecoverPhrase[],
    newPassword: string,
    passwordConfirm: string,
  ) => Promise<void>;
  regenerateWords: (password: string) => Promise<MnemonicResponse>;
  changePassword: (current: string, newPwd: string) => Promise<void>;
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
    if (t) {
      apiProfile()
        .then((p) => {
          const info: UserInfo = {
            username: p.username,
            twofaEnabled: p.twofa_enabled,
            pinCodeSet: p.pincode_set,
          };
          saveUserInfo(info);
          setUserInfo(info);
        })
        .catch(() => {});
    }
  }, []);

  const login = async (
    username: string,
    password: string,
  ): Promise<void> => {
    const t = await apiLogin(username, password);
    saveTokens(t);
    setTokens(t);
    const p = await apiProfile();
    const info: UserInfo = {
      username: p.username,
      twofaEnabled: p.twofa_enabled,
      pinCodeSet: p.pincode_set,
    };
    saveUserInfo(info);
    setUserInfo(info);
  };

  const register = async (
    username: string,
    password: string,
    passwordConfirm: string,
  ): Promise<RegisterResponse> => {
    const res = await apiRegister(username, password, passwordConfirm);
    return res;
  };

  const logout = async (): Promise<void> => {
    try {
      await apiLogout();
    } finally {
      clearTokens();
      setTokens(null);
      clearUserInfo();
      setUserInfo(null);
    }
  };

  const refresh = async (): Promise<void> => {
    const t = await apiRefresh();
    saveTokens(t);
    setTokens(t);
    const p = await apiProfile();
    const info: UserInfo = {
      username: p.username,
      twofaEnabled: p.twofa_enabled,
      pinCodeSet: p.pincode_set,
    };
    saveUserInfo(info);
    setUserInfo(info);
  };

  const recover = async (
    username: string,
    phrases: RecoverPhrase[],
    newPassword: string,
    passwordConfirm: string,
  ): Promise<void> => {
    const t = await apiRecover(
      username,
      phrases,
      newPassword,
      passwordConfirm,
    );
    saveTokens(t);
    setTokens(t);
    const p = await apiProfile();
    const info: UserInfo = {
      username: p.username,
      twofaEnabled: p.twofa_enabled,
      pinCodeSet: p.pincode_set,
    };
    saveUserInfo(info);
    setUserInfo(info);
  };

  const regenerateWords = async (
    password: string,
  ): Promise<MnemonicResponse> => {
    return apiRegenerate(password);
  };

  const changePassword = async (
    current: string,
    newPwd: string,
  ): Promise<void> => {
    await apiChangePassword(current, newPwd);
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
