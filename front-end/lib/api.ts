import { explain } from "./format";

export type Account = {
  id: string;
  label: string;
  created_at: string;
};

export type Balance = {
  asset: "BRL" | "BTC";
  available: string;
  reserved: string;
};

export type Order = {
  id: string;
  account_id: string;
  instrument: string;
  side: "buy" | "sell";
  price: string;
  quantity: string;
  remaining_quantity: string;
  status: "open" | "filled" | "canceled";
  created_at: string;
};

export type Trade = {
  price: string;
  quantity: string;
  maker_order_id: string;
  taker_order_id: string;
  buyer_account_id: string;
  seller_account_id: string;
  created_at?: string;
};

export type Book = {
  instrument: string;
  bids: Order[];
  asks: Order[];
};

export class ApiError extends Error {}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`/backend${path}`, {
    ...init,
    headers: {
      ...(init?.body ? { "content-type": "application/json" } : {}),
      ...(init?.headers ?? {}),
    },
    cache: "no-store",
  });
  const text = await res.text();
  let data: unknown = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch {
      data = { error: text };
    }
  }
  if (!res.ok) {
    const raw =
      typeof data === "object" && data && "error" in data
        ? String((data as { error: unknown }).error)
        : res.statusText;
    throw new ApiError(explain(raw));
  }
  return data as T;
}
