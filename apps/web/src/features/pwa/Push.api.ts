import { apiRequest } from "@/lib/api/client";

export type PushSubscribeResponse = {
  id: string;
  public_id: string;
  endpoint: string;
  user_agent?: string;
  created_at: string;
  updated_at: string;
};

export type VapidPublicKeyResponse = {
  public_key: string;
};

export async function fetchVapidPublicKey(): Promise<string> {
  const data = await apiRequest<VapidPublicKeyResponse>(
    "/api/v1/push/vapid-public-key",
  );
  return data.public_key;
}

export async function subscribePush(input: {
  endpoint: string;
  p256dh: string;
  auth: string;
}): Promise<PushSubscribeResponse> {
  return apiRequest<PushSubscribeResponse>("/api/v1/push/subscribe", {
    method: "POST",
    body: {
      endpoint: input.endpoint,
      keys: {
        p256dh: input.p256dh,
        auth: input.auth,
      },
    },
  });
}

export async function unsubscribePush(endpoint: string): Promise<void> {
  await apiRequest<{ unsubscribed: boolean }>("/api/v1/push/unsubscribe", {
    method: "DELETE",
    body: { endpoint },
  });
}
