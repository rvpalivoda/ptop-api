import { apiRequest } from "./client";
import { loadTokens } from "@/storage/token";

interface AuthResponse {
  access: string;
  refresh: string;
}

interface RegisterResponse {
  seed: string;
}

export async function login(
  username: string,
  password: string,
  captcha: string,
) {
  return apiRequest<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ username, password, captcha }),
  });
}

export async function register(
  username: string,
  password: string,
  captcha: string,
): Promise<RegisterResponse> {
  return apiRequest<RegisterResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, password, captcha }),
  });
}

export async function recover(
  username: string,
  words: string[],
  indices: number[],
  captcha: string,
  newPassword?: string,
) {
  const payload: Record<string, unknown> = {
    username,
    words,
    indices,
    captcha,
  };
  if (newPassword) payload.new_password = newPassword;
  return apiRequest<AuthResponse>("/auth/recover", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function regenerateWords(password: string) {
  return apiRequest<RegisterResponse>("/auth/regenerate", {
    method: "POST",
    body: JSON.stringify({ password }),
  });
}

export async function changePassword(currentPassword: string, newPassword: string) {
  return apiRequest<
    { status: string }
  >("/auth/change-password", {
    method: "POST",
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
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
  return apiRequest<AuthResponse>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh: tokens.refresh }),
  });
}

export async function logout() {
  const tokens = loadTokens();
  await apiRequest("/auth/logout", {
    method: "POST",
    body: JSON.stringify({ refresh: tokens?.refresh }),
  });
}
