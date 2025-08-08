import { apiRequest } from "./client";
import { loadTokens } from "@/storage/token";

interface TokenResponse {
  access_token: string;
  refresh_token: string;
}

export interface MnemonicWord {
  position: number;
  word: string;
}

export interface RegisterResponse {
  access: string;
  refresh: string;
  mnemonic: MnemonicWord[];
}

export interface MnemonicResponse {
  mnemonic: string;
}

export async function login(
  username: string,
  password: string,
) {
  const { access_token, refresh_token } = await apiRequest<TokenResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password }),
  });
  return { access: access_token, refresh: refresh_token };
}

export async function register(
  username: string,
  password: string,
  passwordConfirm: string,
): Promise<RegisterResponse> {
  const { access_token, refresh_token, mnemonic } = await apiRequest<
    TokenResponse & { mnemonic: MnemonicWord[] }
  >("/auth/register", {
    method: "POST",
    body: JSON.stringify({
      username,
      password,
      password_confirm: passwordConfirm,
    }),
  });
  return { access: access_token, refresh: refresh_token, mnemonic };
}

export interface RecoverPhrase {
  position: number;
  word: string;
}

export async function recoverChallenge(username: string) {
  return apiRequest<{ positions: number[] }>(`/auth/recover/${username}`);
}

export async function recover(
  username: string,
  phrases: RecoverPhrase[],
  newPassword: string,
  passwordConfirm: string,
) {
  const payload = {
    username,
    phrases,
    new_password: newPassword,
    password_confirm: passwordConfirm,
  };
  const { access_token, refresh_token } = await apiRequest<TokenResponse>(
    "/auth/recover",
    {
      method: "POST",
      body: JSON.stringify(payload),
    },
  );
  return { access: access_token, refresh: refresh_token };
}

export async function regenerateWords(password: string) {
  return apiRequest<MnemonicResponse>("/auth/regenerate", {
    method: "POST",
    body: JSON.stringify({ password }),
  });
}

export async function changePassword(
  currentPassword: string,
  newPassword: string,
) {
  return apiRequest<{ status: string }>("/auth/password", {
    method: "POST",
    body: JSON.stringify({
      old_password: currentPassword,
      new_password: newPassword,
      confirm_password: newPassword,
    }),
  });
}

export async function verifyPassword(password: string) {
  return apiRequest<{ verified: boolean }>("/auth/verify-password", {
    method: "POST",
    body: JSON.stringify({ password }),
  });
}

export async function refresh() {
  const tokens = loadTokens();
  if (!tokens?.refresh) throw new Error("No refresh token");
  const { access_token, refresh_token } = await apiRequest<TokenResponse>(
    "/auth/refresh",
    {
      method: "POST",
      body: JSON.stringify({ refresh_token: tokens.refresh }),
    },
  );
  return { access: access_token, refresh: refresh_token };
}

export interface ProfileResponse {
  username: string;
  twofa_enabled: boolean;
  pincode_set: boolean;
}

export async function profile() {
  return apiRequest<ProfileResponse>("/auth/profile");
}

export interface StatusResponse {
  status: string;
}

export async function logout() {
  return apiRequest<StatusResponse>("/auth/logout", {
    method: "POST",
  });
}
