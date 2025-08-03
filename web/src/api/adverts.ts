import { apiRequest } from "./client";

export interface PaymentMethodPayload {
  payment_method: string;
  payment_method_name: string;
}

export interface CreateAdvertPayload {
  client: string;
  status: string;
  type: "Buy" | "Sell";
  base_asset: string;
  quote_asset: string;
  amount: number;
  price: number;
  country: string;
  post_code: string;
  payment_method: PaymentMethodPayload[];
  visibility_status?: "public" | "private" | "private-by-link";
  terms?: string;
  min_amount?: number;
  max_amount?: number;
  duration: string;
}

interface CreateAdvertResponse {
  advert_id: string;
}

export async function createAdvert(
  data: CreateAdvertPayload,
): Promise<CreateAdvertResponse> {
  return apiRequest<CreateAdvertResponse>("/adverts", {
    method: "POST",
    body: JSON.stringify(data),
  });
}
