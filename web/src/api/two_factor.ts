import { apiRequest } from "./client";

interface EnableResponse {
  secret: string;
  url: string;
}

export async function enable2fa(password: string) {
  return apiRequest<EnableResponse>("/auth/2fa/enable", {
    method: "POST",
    body: JSON.stringify({ password }),
  });
}

export async function disable2fa(password: string) {
  return apiRequest("/auth/2fa/disable", {
    method: "POST",
    body: JSON.stringify({ password }),
  });
}
